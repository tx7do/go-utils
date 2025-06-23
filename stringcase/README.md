# 命名法转换

- camelCase 驼峰式命名法（大驼峰）
- PascalCase 帕斯卡命名法（小驼峰）
- snake_case 蛇形命名法
- kebab-case 烤肉串命名法

## camelCase 驼峰式命名法（大驼峰）

驼峰式命名法（Camel case）是一种不使用空格，将多个单词连起来形成一个标识符的命名方式，其中每个单词的首字母（除了第一个单词，如果使用小驼峰式命名法）都大写，就像骆驼的驼峰一样。

驼峰式命名法分为两种：首字母小写的“小驼峰式”（lowerCamelCase）和首字母大写的“大驼峰式”（UpperCamelCase，也称为帕斯卡命名法PascalCase）。

- **小驼峰式(lowerCamelCase)**: 第一个单词的首字母小写，后续单词的首字母大写。例如：`myVariableName`。
- **大驼峰式(UpperCamelCase)**: 每个单词的首字母都大写。例如：MyVariableName，也称为帕斯卡命名法（PascalCase）。

在 JavaScript、Java和C#中，驼峰式大小写常用于变量和函数的命名。

```javascript
let firstName = "John";
let lastName = "Doe";

function printFullName(firstName, lastName) {
    let fullName = firstName + " " + lastName;
    console.log(fullName);
}
```

## PascalCase 帕斯卡命名法（小驼峰）

`PascalCase`，也称为`UpperCamelCase`，是一种在编程中使用的命名约定。它要求每个单词（包括第一个单词）的首字母都大写，并且单词之间没有空格或分隔符（如
`_`）。例如，`ThisIsPascalCase`，`MyClassName` 都是使用PascalCase的例子。

PascalCase 通常用于在 C#、Java 和TypeScript等语言中命名类、接口和其他类型。

```typescript
class Person {
    firstName: string;
    lastName: string;

    constructor(firstName: string, lastName: string) {
        this.firstName = firstName;
        this.lastName = lastName;
    }

    printFullName(): void {
        let fullName = this.firstName + " " + this.lastName;
        console.log(fullName);
    }
}
```

## snake_case 蛇形命名法

蛇形命名法是一种使用下划线 (`_`) 分隔单词的命名方式。之所以叫蛇形命名法，是因为"`snake_case`"的
下划线的形状类似于蛇腹上的鳞片。蛇形命名法通常用于Python、Ruby 和JavaScript等语言的变量名和函数名。

```python
first_name = "John"
last_name = "Doe"

def print_full_name(first_name, last_name):
    full_name = first_name + " " + last_name
    print(full_name)
```

## kebab-case 烤肉串命名法

`kebab-case（烤肉串命名法）`，也被称作 `kebab case`、`dash-case（破折号式）`、`hyphen-case（连字符式）`、`lisp-case（Lisp 式）`。

kebab-case 要求短语内的各个单词或缩写之间以`-`（连字符）做间隔。 例如："`kebab-case`"。

短横线命名法通常用于 URL、文件名和 HTML/CSS 类名。

```html

<div class="user-profile">
    <p>This is a user profile.</p>
</div>
```
