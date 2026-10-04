# fieldperm

字段级权限的通用消息裁剪原语，全部基于 protoreflect 反射实现，新增受控资源无需改执行器。

隐藏字段集来源于令牌 claim（登录期按角色并集聚合，条目格式 `"资源.字段"`，
字段为 proto json_name），执行侧用它对 proto 消息做读写两向裁剪：

- **读路径**（`ApplyReadMask` / `ApplyReadMaskList`）：就地清除命中字段。protojson 默认不输出
  未填充字段，清值即等于从 JSON 响应中移除该字段。
- **写路径**（`StripWriteFields`）：清除请求载荷中命中字段的值，并同步剔除
  field_mask 中的对应路径，防止 Update 的按 mask 过滤语义把已清空的值当作显式置零写库。

```go
import "github.com/tx7do/go-utils/fieldperm"

hidden := fieldperm.ParseTokenEntries([]string{"User.passwordHash", "User.salt"})
fields := fieldperm.HiddenFieldsOf(entries, "User")
fieldperm.ApplyReadMask(msg, fields)
fieldperm.StripWriteFields(req.Data, req.UpdateMask, fields)
```
