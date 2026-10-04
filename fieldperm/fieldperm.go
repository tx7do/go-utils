// Package fieldperm 提供字段级权限的通用消息裁剪原语。
//
// 隐藏字段集来源于令牌 claim（登录期按角色并集聚合，条目格式 "资源.字段"，
// 字段为 proto json_name），执行侧用它对 proto 消息做读写两向裁剪：
//
//   - 读路径（ApplyReadMask）：就地清除命中字段。protojson 默认不输出
//     未填充字段，清值即等于从 JSON 响应中移除该字段。
//   - 写路径（StripWriteFields）：清除请求载荷中命中字段的值，并同步剔除
//     field_mask 中的对应路径，防止 Update 的 FilterByFieldMask 语义把
//     已清空的值当作显式置零写库。
//
// 全部基于 protoreflect 反射实现，新增受控资源无需改执行器。
package fieldperm

import (
	"strings"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

// ParseTokenEntries 把令牌隐藏字段条目（"资源.字段" 串）解析为
// 资源名 → 字段名集合。字段集合同时收录 json_name 与 proto 字段名两种拼写：
// 配置/展示侧用 json_name（camelCase），而 field_mask 路径经 protojson
// 规范化后一律是 proto 字段名（snake_case），两侧都必须能命中。
func ParseTokenEntries(entries []string) map[string]map[string]struct{} {
	result := make(map[string]map[string]struct{})
	for _, entry := range entries {
		resource, field, ok := strings.Cut(entry, ".")
		if !ok || resource == "" || field == "" {
			continue
		}
		fields, ok := result[resource]
		if !ok {
			fields = make(map[string]struct{})
			result[resource] = fields
		}
		fields[field] = struct{}{}
	}
	return result
}

// HiddenFieldsOf 返回指定资源在令牌条目中的隐藏字段集；无配置返回 nil。
func HiddenFieldsOf(entries []string, resource string) map[string]struct{} {
	parsed := ParseTokenEntries(entries)
	return parsed[resource]
}

// IsEmpty 判断隐藏字段集是否为空（空集 = 无需裁剪）。
func IsEmpty(hidden map[string]struct{}) bool {
	return len(hidden) == 0
}

// ApplyReadMask 就地清除消息顶层命中隐藏集的字段（含 json_name/proto 名双拼写命中）。
// 嵌套消息字段不做递归——本原语只管顶层业务字段。
func ApplyReadMask(msg proto.Message, hidden map[string]struct{}) {
	if msg == nil || IsEmpty(hidden) {
		return
	}

	m := msg.ProtoReflect()
	if !m.IsValid() {
		return
	}

	fields := m.Descriptor().Fields()
	for i := 0; i < fields.Len(); i++ {
		fd := fields.Get(i)
		if _, ok := hidden[string(fd.Name())]; ok {
			m.Clear(fd)
			continue
		}
		if jsonName := fd.JSONName(); jsonName != string(fd.Name()) {
			if _, ok := hidden[jsonName]; ok {
				m.Clear(fd)
			}
		}
	}
}

// ApplyReadMaskList 对消息列表逐个应用读裁剪（List 响应的 items）。
func ApplyReadMaskList[T proto.Message](items []T, hidden map[string]struct{}) {
	if IsEmpty(hidden) {
		return
	}
	for _, item := range items {
		ApplyReadMask(item, hidden)
	}
}

// StripWriteFields 剥离写请求载荷中的隐藏字段：
//   - 清除 data 消息中命中字段的值（ent 侧 nil-skip 语义下等于"忽略该字段"）；
//   - 从 field_mask 路径中剔除命中项（proto 名与 json_name 都要剔除），
//     防止"mask 声明了该字段但值为零"被当成显式清空写库。
func StripWriteFields(data proto.Message, fieldMask *fieldmaskpb.FieldMask, hidden map[string]struct{}) {
	if IsEmpty(hidden) {
		return
	}

	ApplyReadMask(data, hidden)

	if fieldMask != nil && len(fieldMask.Paths) > 0 {
		kept := fieldMask.Paths[:0]
		for _, p := range fieldMask.Paths {
			// field_mask 路径可能是 "field" 或 "field.sub"——顶层段命中即整条剔除。
			top := p
			if idx := strings.IndexByte(p, '.'); idx >= 0 {
				top = p[:idx]
			}
			if _, ok := hidden[top]; ok {
				continue
			}
			kept = append(kept, p)
		}
		fieldMask.Paths = kept
	}
}
