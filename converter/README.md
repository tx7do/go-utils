# converter

权限码转换器：把 API 操作标识映射为 `resource:action` 风格的权限码。

- **`ApiPermissionConverter`**：
  - `ConvertCodeByOperationID(operationID)` — 从 kratos operationID
    （`<Service>_<Method>` 形态）解析出资源与动作；
  - `ConvertCodeByPath(method, path)` — 从 HTTP 方法与 REST 路径解析
    资源与动作（剥离版本前缀与路径参数、折叠多级路径为首段、单数化）。

动作统一归一为 `view/create/edit/delete` 四类。资源名经
[`stringcase`](../stringcase) 归一为 kebab-case。

```go
import "github.com/tx7do/go-utils/converter"

conv := converter.NewApiPermissionConverter()
code := conv.ConvertCodeByOperationID("UserService_ListUsers")  // "user:view"
code = conv.ConvertCodeByPath("GET", "/api/v1/users/{id}")      // "user:view"
```
