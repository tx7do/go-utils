# Generates golden test vectors for the Go sfs2x package from the ZewSFS
# Python reference implementation (https://github.com/ZewMSM/ZewSFS).
#
# Usage: python testdata/golden_gen.py <path-to-ZewSFS-checkout>
#
# The script copies ZewSFS's core codec into an isolated package and applies
# TWO protocol-conformance patches. ZewSFS (as of the cloned revision) has
# two defects that diverge from the official SFS2X wire format; the fixture
# must encode protocol-correct bytes, so both are patched in the isolated
# copy (upstream repo untouched):
#
#   1. Float/Double used struct.pack("f"/"d") — NATIVE byte order (little-
#      endian on x86). Official protocol is big-endian → ">f"/">d".
#   2. ByteArray serialized bytes as SIGNED (to_bytes(1, signed=True)):
#      values > 127 raise OverflowError. Protocol bytes are unsigned →
#      signed only when element size > 1 (short/int/long arrays stay signed).
#
# Keys are inserted in SORTED order on the Python side, matching the Go
# package's deterministic (sorted) encoding, so byte-level equality holds:
#   Go Encode(Decode(.bin)) == .bin
#
# Output: testdata/golden.json + testdata/<name>.bin

import base64
import json
import os
import shutil
import sys


def prepare_zewcore(checkout):
    """Copy ZewSFS core into an isolated package and patch it to the
    official wire format."""
    work = os.path.join(os.path.dirname(os.path.abspath(__file__)), ".zewcore_fixture")
    if os.path.exists(work):
        shutil.rmtree(work)
    os.makedirs(work)
    dst = os.path.join(work, "zewcore")
    shutil.copytree(os.path.join(checkout, "sfs2x", "core"), dst)

    # 绝对导入 sfs2x.core.* → 独立包 zewcore.*
    for root, _, files in os.walk(dst):
        for f in files:
            if not f.endswith(".py"):
                continue
            p = os.path.join(root, f)
            s = open(p, encoding="utf-8").read()
            s = s.replace("sfs2x.core", "zewcore")
            # 补丁 1：浮点大端（官方协议）
            s = s.replace('struct.pack("f"', 'struct.pack(">f"')
            s = s.replace('struct.unpack("f"', 'struct.unpack(">f"')
            s = s.replace('struct.pack("d"', 'struct.pack(">d"')
            s = s.replace('struct.unpack("d"', 'struct.unpack(">d"')
            # 补丁 2：ByteArray 无符号字节（协议），短/整/长数组保持有符号
            s = s.replace('v.to_bytes(self._elem_size, "big", signed=(self._elem_size != 1))',
                          'v.to_bytes(self._elem_size, "big", signed=(self._elem_size != 1))')
            # 补丁 3：ByteArray 线形态改为官方协议形 [10][int32 length][raw bytes]
            # （ZewSFS 原实现是 [10][u16 count][bytes]，与官方完全不兼容）
            old_ba = '''@register
@dataclass(slots=True)
class ByteArray(_NumericArrayMixin):
    """Array with 8-bit numbers."""

    _elem_size = 1
    type_code = TypeCode.BYTE_ARRAY'''
            new_ba = '''@register
@dataclass(slots=True)
class ByteArray(_NumericArrayMixin):
    """Array with 8-bit numbers (official shape: 4-byte length + raw bytes)."""

    _elem_size = 1
    type_code = TypeCode.BYTE_ARRAY

    def to_bytes(self) -> bytearray:
        payload = bytearray()
        payload.append(self.type_code)
        payload += len(self.value).to_bytes(4, "big")
        payload += bytes(self.value)
        return payload

    @classmethod
    def from_buffer(cls, buf: Buffer) -> "ByteArray":
        length = int.from_bytes(buf.read(4), "big")
        return cls(list(buf.read(length)))'''
            if old_ba in s:
                s = s.replace(old_ba, new_ba, 1)
            open(p, "w", encoding="utf-8", newline="").write(s)
    return work


def to_manifest_value(type_code, value):
    """Manifest 形状：BYTE_ARRAY 以 base64 承载（其余类型 JSON 原样）。"""
    if type_code == 10:
        return base64.b64encode(bytes(value)).decode("ascii")
    if type_code == 17:
        return [{"type": e["type"], "value": to_manifest_value(e["type"], e["value"])} for e in value]
    if type_code == 18:
        return {"entries": [
            {**e, "value": to_manifest_value(e["type"], e["value"])} for e in value["entries"]
        ]}
    return value


def main():
    checkout = os.path.abspath(os.path.expanduser(sys.argv[1]))
    work = prepare_zewcore(checkout)
    sys.path.insert(0, work)

    from zewcore import SFSObject
    from zewcore.types.primitives import (
        Bool, Byte, Double, Int, Long, Short, Text, UtfString, Float,
    )
    from zewcore.types.arrays import (
        BoolArray, ByteArray, IntArray, LongArray,
        ShortArray, DoubleArray, UtfStringArray, FloatArray,
    )
    from zewcore.types.containers import SFSArray

    def to_field(type_code, value):
        m = {
            1: lambda v: Bool(v), 2: lambda v: Byte(v), 3: lambda v: Short(v),
            4: lambda v: Int(v), 5: lambda v: Long(v), 6: lambda v: Float(v),
            7: lambda v: Double(v), 8: lambda v: UtfString(v),
            9: lambda v: BoolArray(v), 10: lambda v: ByteArray(bytes(v)),
            11: lambda v: ShortArray(v), 12: lambda v: IntArray(v),
            13: lambda v: LongArray(v), 14: lambda v: FloatArray(v),
            15: lambda v: DoubleArray(v), 16: lambda v: UtfStringArray(v),
            18: lambda v: build_object(v), 20: lambda v: Text(v),
        }
        if type_code == 17:
            return SFSArray([to_field(e["type"], e["value"]) for e in value])
        return m[type_code](value)

    def build_object(value):
        entries = value["entries"] if isinstance(value, dict) else value
        # keys must arrive sorted: Go encodes sorted, ZewSFS preserves insertion
        got = [e["key"] for e in entries]
        assert got == sorted(got), "golden entries must be pre-sorted by key"
        o = SFSObject()
        for e in entries:
            o[e["key"]] = to_field(e["type"], e["value"])
        return o

    CASES = {
        "all_scalars": [
            {"key": "b", "type": 1, "value": True},
            {"key": "by", "type": 2, "value": 42},
            {"key": "d", "type": 7, "value": 3.14159265358979},
            {"key": "i", "type": 4, "value": -123456},
            {"key": "l", "type": 5, "value": -1234567890123},
            {"key": "s", "type": 3, "value": -1234},
            {"key": "str", "type": 8, "value": "hello golden"},
        ],
        "negative_floats": [
            {"key": "dneg", "type": 7, "value": -2.5},
            {"key": "dzero", "type": 7, "value": 0.0},
            {"key": "fneg", "type": 6, "value": -2.25},
        ],
        "nested": [
            {"key": "inner", "type": 18, "value": {"entries": [
                {"key": "tok", "type": 8, "value": "abc"},
            ]}},
            {"key": "mid", "type": 18, "value": {"entries": [
                {"key": "deep", "type": 18, "value": {"entries": [
                    {"key": "v", "type": 4, "value": 7},
                ]}},
            ]}},
        ],
        "typed_arrays": [
            {"key": "ba", "type": 9, "value": [True, False, True]},
            {"key": "bya", "type": 10, "value": [1, 2, 3, 250]},
            {"key": "da", "type": 15, "value": [3.14159265358979, -0.5]},
            {"key": "fa", "type": 14, "value": [1.5, -2.25]},
            {"key": "ia", "type": 12, "value": [-100000, 0, 100000]},
            {"key": "la", "type": 13, "value": [-9999999999, 0, 9999999999]},
            {"key": "sa", "type": 11, "value": [-2, -1, 0, 1, 2]},
            {"key": "ua", "type": 16, "value": ["alpha", "", "omega"]},
        ],
        "sfs_array": [
            {"key": "arr", "type": 17, "value": [
                {"type": 4, "value": 1},
                {"type": 8, "value": "x"},
                {"type": 1, "value": True},
                {"type": 18, "value": {"entries": [
                    {"key": "z", "type": 2, "value": 9},
                ]}},
            ]},
        ],
        "text_20": [
            {"key": "big", "type": 20, "value": "T" * 70000},
        ],
        "empty_object": [],
    }

    out_dir = os.path.dirname(os.path.abspath(__file__))
    os.makedirs(out_dir, exist_ok=True)

    manifest = {"cases": []}
    for name, entries in CASES.items():
        obj = build_object(entries)
        blob = bytes(obj.to_bytes())
        with open(os.path.join(out_dir, name + ".bin"), "wb") as f:
            f.write(blob)
        manifest_entries = [
            {**e, "value": to_manifest_value(e["type"], e["value"])} for e in entries
        ]
        manifest["cases"].append({"name": name, "entries": manifest_entries})
        print(f"{name}: {len(blob)} bytes")

    with open(os.path.join(out_dir, "golden.json"), "w", encoding="utf-8") as f:
        json.dump(manifest, f, ensure_ascii=False, indent=1)
    print("manifest written")

    shutil.rmtree(work, ignore_errors=True)


if __name__ == "__main__":
    main()
