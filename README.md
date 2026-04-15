### 介绍
这是封装一些`go`开发中常用的工具方法,使用方式
* 1、安装依赖包
    ```protobuf
    go get -u github.com/kuangshp/go-utils
    ```

---

## 目录

- [验证器 (Validator)](#验证器-validator)
- [字符串操作 (String)](#字符串操作-string)
- [类型转换 (Conv)](#类型转换-conv)
- [Map 操作 (Map)](#map-操作-map)
- [Slice 操作 (Slice)](#slice-操作-slice)
- [切片链式调用 (Builder)](#切片链式调用-builder)
- [树形结构 (Tree)](#树形结构-tree)
- [时间处理 (Time)](#时间处理-time)
- [随机数 (Random)](#随机数-random)
- [重试机制 (Retry)](#重试机制-retry)
- [HTTP 客户端 (HttpClient)](#http-客户端-httpclient)
- [文件目录 (Folder)](#文件目录-folder)
- [Excel 导出 (Excel)](#excel-导出-excel)
- [验证码 (Captcha)](#验证码-captcha)
- [图片缓存存储 (Store)](#图片缓存存储-store)
- [身份证操作 (Card)](#身份证操作-card)
- [IP 地址操作 (IP)](#ip-地址操作-ip)
- [URL 操作 (URL)](#url-操作-url)
- [条件工具 (If)](#条件工具-if)

---

## 验证器 (Validator)

位于 `k/validator.go`

### 字符串验证

| 函数 | 说明 | 示例 |
|------|------|------|
| `IsAlpha(str)` | 判断是否为字母(大写/小写) | `IsAlpha("abc")` → true |
| `IsAllUpper(str)` | 判断是否全部为大写字母 | `IsAllUpper("ABC")` → true |
| `IsAllLower(str)` | 判断是否全部为小写字母 | `IsAllLower("abc")` → true |
| `IsASCII(str)` | 判断字符串是否全部为ASCII | `IsASCII("hello")` → true |
| `ContainUpper(str)` | 判断是否包含大写字母 | `ContainUpper("abcD")` → true |
| `ContainLower(str)` | 判断是否包含小写字符 | `ContainLower("ABCd")` → true |
| `ContainLetter(str)` | 判断是否至少包含一个字母 | `ContainLetter("123")` → false |
| `ContainNumber(input)` | 判断是否至少包含一个数字 | `ContainNumber("abc1")` → true |
| `IsEmptyString(str)` | 判断是否为空字符串 | `IsEmptyString("")` → true |

### 格式验证

| 函数 | 说明 | 示例 |
|------|------|------|
| `IsJSON(str)` | 判断是否为JSON格式 | `IsJSON("{\"key\":\"value\"}")` → true |
| `IsUrl(str)` | 判断是否为URL地址 | `IsUrl("https://example.com")` → true |
| `IsDns(dns)` | 判断是否为DNS格式 | `IsDns("example.com")` → true |
| `IsEmail(email)` | 判断是否为Email地址 | `IsEmail("test@example.com")` → true |
| `IsBase64(base64)` | 判断是否为Base64字符串 | `IsBase64("SGVsbG8=")` → true |

### 数字验证

| 函数 | 说明 | 示例 |
|------|------|------|
| `IsNumberStr(s)` | 判断是否为数字字符串 | `IsNumberStr("123")` → true |
| `IsFloatStr(str)` | 判断是否为浮点型字符串 | `IsFloatStr("3.14")` → true |
| `IsIntStr(str)` | 判断是否为整型字符串 | `IsIntStr("123")` → true |
| `IsRegexMatch(str, regex)` | 判断是否匹配正则 | `IsRegexMatch("abc", "^[a-z]+$")` → true |

### IP/端口验证

| 函数 | 说明 | 示例 |
|------|------|------|
| `IsIp(ipStr)` | 判断是否为IP地址 | `IsIp("192.168.1.1")` → true |
| `IsIpV4(ipStr)` | 判断是否为IPv4 | `IsIpV4("192.168.1.1")` → true |
| `IsIpV6(ipStr)` | 判断是否为IPv6 | `IsIpV6("::1")` → true |
| `IsPort(str)` | 判断是否为有效端口号 | `IsPort("8080")` → true |

### 日期验证

| 函数 | 说明 | 示例 |
|------|------|------|
| `IsDate(date)` | 判断是否为日期格式 YYYY-MM-DD | `IsDate("2024-01-01")` → true |
| `IsDateTime(dateTime)` | 判断是否为日期时间格式 | `IsDateTime("2024-01-01 12:00:00")` → true |

### 中国相关验证

| 函数 | 说明 | 示例 |
|------|------|------|
| `IsChineseMobile(mobileNum)` | 判断是否为中国手机号 | `IsChineseMobile("13800138000")` → true |
| `IsChineseIdNum(id)` | 判断是否为中国大陆身份证号 | `IsChineseIdNum("110101199001011234")` → true |
| `IsContainChinese(s)` | 判断是否包含中文 | `IsContainChinese("你好")` → true |
| `IsChinesePhone(phone)` | 判断是否为中国电话号码 | `IsChinesePhone("010-12345678")` → true |

### 银行卡验证

| 函数 | 说明 | 示例 |
|------|------|------|
| `IsCreditCard(creditCart)` | 判断是否为信用卡号 | `IsCreditCard("4111111111111111")` → true |

### 通用验证

| 函数 | 说明 | 示例 |
|------|------|------|
| `IsZeroValue(value)` | 判断是否为空值(nil/零值) | `IsZeroValue(0)` → true |

---

## 字符串操作 (String)

位于 `k/string.go`

| 函数 | 说明 | 示例 |
|------|------|------|
| `Substring(s, offset, length)` | 字符串截取 | `Substring("hello", 0, 3)` → "hel" |
| `HideString(origin, start, end, replace)` | 隐藏字符串部分字符 | `HideString("13800138000", 3, 7, "*")` → "138****8000" |
| `MaskEmail(email)` | 隐藏邮箱中间4位 | `MaskEmail("test@example.com")` → "te****@example.com" |
| `MaskMobile(mobile)` | 隐藏手机号中间4位 | `MaskMobile("13800138000")` → "138****8000" |
| `MakePassword(password)` | 明文转换为bcrypt密文 | `MakePassword("123456")` → "$2a$10$..." |
| `CheckPassword(encrypted, password)` | 校验密码是否正确 | `CheckPassword(hash, "123456")` → true |
| `JoinStr(sep, parts...)` | 字符串拼接(过滤空字符串) | `JoinStr("-", "a", "", "b")` → "a-b" |
| `Case2Camel(name)` | 下划线转大驼峰 | `Case2Camel("hello_world")` → "HelloWorld" |
| `LowerCamelCase(name)` | 下划线转小驼峰 | `LowerCamelCase("hello_world")` → "helloWorld" |

---

## 类型转换 (Conv)

位于 `k/conv.go`

| 函数 | 说明 | 示例 |
|------|------|------|
| `StringToInt(str)` | string转int | `StringToInt("123")` → 123 |
| `StringToUint(str)` | string转uint | `StringToUint("123")` → 123 |
| `StringToInt64(str)` | string转int64 | `StringToInt64("123")` → 123 |
| `IntToString(num)` | int转string | `IntToString(123)` → "123" |
| `UintToString(num)` | uint转string | `UintToString(123)` → "123" |
| `Int64ToString(num)` | int64转string | `Int64ToString(123)` → "123" |

---

## Map 操作 (Map)

位于 `k/map.go`

| 函数 | 说明 | 示例 |
|------|------|------|
| `MapToString(result)` | map转换为JSON字符串 | `MapToString(map[string]any{"a":1})` → `{"a":1}` |
| `MapIndentToString(result)` | map转换为格式化JSON字符串 | `MapIndentToString(...)` → 格式化JSON |
| `MapKeySort(m)` | map的key按ASCII排序并拼接 | `MapKeySort(map[string]int{"b":1,"a":2})` → "a=2&b=1" |
| `Keys(m)` | 获取map的全部key | `Keys(map[string]int{"a":1})` → ["a"] |
| `Values(m)` | 获取map的全部value | `Values(map[string]int{"a":1})` → [1] |
| `KeysBy(m, mapper)` | 根据条件获取map的key | `KeysBy(m, func(k string) int{...})` → [] |
| `ValuesBy(m, mapper)` | 根据条件获取map的值 | `ValuesBy(m, func(v any) int{...})` → [] |
| `Merge(maps...)` | 合并多个map | `Merge(map1, map2)` → 合并后的map |
| `ForEachMap(m, iteratee)` | 循环遍历map | `ForEachMap(m, func(k,v){...})` |
| `FilterMap(m, predicate)` | 过滤map | `FilterMap(m, func(k,v)bool{...})` → 过滤后的map |
| `FilterByKeys(m, keys)` | 根据key过滤数据 | `FilterByKeys(m, ["a","b"])` → 只包含指定key的map |
| `FilterByValues(m, values)` | 根据value过滤数据 | `FilterByValues(m, [1,2])` → 只包含指定值的map |
| `OmitBy(m, predicate)` | 删除满足条件的元素 | `OmitBy(m, func(k,v)bool{...})` |
| `OmitByKeys(m, keys)` | 根据key删除元素 | `OmitByKeys(m, ["a"])` → 删除指定key的map |
| `OmitByValues(m, values)` | 根据value删除元素 | `OmitByValues(m, [1])` → 删除指定值的map |
| `MapKeys(m, iteratee)` | 抽取map的key生成新map | `MapKeys(m, func(k,v)string{...})` |
| `MapValues(m, iteratee)` | 抽取map的value生成新map | `MapValues(m, func(k,v)int{...})` |
| `HasKey(m, key)` | 判断map是否包含某个key | `HasKey(m, "a")` → true/false |
| `GetValue(data, key, defaultValue)` | 获取map值，支持默认值 | `GetValue(m, "a", 0)` → 值或默认值 |
| `StructToMap(obj)` | 结构体转map(使用json标签) | `StructToMap(User{})` → map[string]any |
| `MapToStruct(m, obj)` | map转结构体 | `MapToStruct(m, &user)` |
| `AnyToStruct(data, obj)` | 任意类型转结构体 | `AnyToStruct(data, &user)` |
| `InterfaceToString(value)` | interface转字符串 | `InterfaceToString(123)` → "123" |

---

## Slice 操作 (Slice)

位于 `k/slice.go`

| 函数 | 说明 | 示例 |
|------|------|------|
| `SortListMap(collection, predicate)` | 对结构体切片排序 | `SortListMap(users, func(u1,u2 User)bool{return u1.Age < u2.Age})` |
| `IsContains(slice, target)` | 判断元素是否在切片中 | `IsContains([]int{1,2,3}, 2)` → true |
| `ForEach(collection, iteratee)` | 遍历切片 | `ForEach(slice, func(item, idx){...})` |
| `Map(collection, iteratee)` | 遍历生成新切片 | `Map([]int{1,2}, func(i int)string{return strconv.Itoa(i)})` → ["1","2"] |
| `Every(slice, predicate)` | 判断所有元素是否满足条件 | `Every([]int{1,2,3}, func(i int)bool{return i>0})` → true |
| `Some(slice, predicate)` | 判断是否有元素满足条件 | `Some([]int{1,2,3}, func(i int)bool{return i>2})` → true |
| `Filter(collection, predicate)` | 过滤切片 | `Filter([]int{1,2,3}, func(i int)bool{return i>1})` → [2,3] |
| `Reduce(collection, accumulator, initial)` | 聚合操作 | `Reduce([]int{1,2,3}, func(sum,i int)int{return sum+i}, 0)` → 6 |
| `ReduceBy(slice, initial, reducer)` | 带初始值的聚合 | 同上 |
| `ReduceRight(collection, accumulator, initial)` | 从右向左聚合 | `ReduceRight([]int{1,2,3}, func(sum,i int)int{return sum-i}, 0)` → 0 |
| `GroupBy(collection, iteratee)` | 分组 | `GroupBy(users, func(u User)string{return u.Dept})` → map[string][]User |
| `GroupByWithMapper(items, keyMapper, valueMapper)` | 带映射的分组 | `GroupByWithMapper(items, func(t T)K{...}, func(t T)V{...})` |
| `Difference(slice, comparedSlice)` | 取差集 | `Difference([1,2,3,4], [1,2])` → [3,4] |
| `Intersect(slice1, slice2)` | 取交集 | `Intersect([1,2,3], [2,3,4])` → [2,3] |
| `Union(slices...)` | 取并集 | `Union([1,2], [2,3])` → [1,2,3] |
| `Distinct(slice)` | 去重 | `Distinct([1,1,2,3])` → [1,2,3] |
| `ToMap(collection, transform)` | 切片转map | `ToMap(users, func(u User)(u.Id, u.Name))` → map[id]name |
| `Index(slice, element)` | 查找元素索引位置 | `Index([1,2,3], 2)` → 1, 不存在返回-1 |
| `Find(slice, predicate)` | 查找第一个满足条件的元素 | `Find([]int{1,2,3}, func(i int)bool{return i>1})` → (2, true) |
| `DistinctByField(slice, fieldName)` | 根据字段去重 | `DistinctByField(users, "Name")` |
| `ChunkSlice(list, size)` | 将切片分块 | `ChunkSlice([1,2,3,4,5], 2)` → [[1,2],[3,4],[5]] |
| `Flatten(b)` | 扁平化一层 | `Flatten(builder)` |

---

## 切片链式调用 (Builder)

位于 `k/slice.go`

流式API风格的切片处理，链式调用。

```go
// 创建
builder := k.From([]int{3, 1, 2})

// 链式操作
result := k.From([]int{3, 1, 2, 4, 5, 6}).
    Filter(func(item int, idx int) bool { return item > 2 }).
    Sort(func(a, b int) bool { return a < b }).
    Take(3).
    Build()
// result: [3, 4, 5]
```

### 支持的方法

| 方法 | 说明 | 示例 |
|------|------|------|
| `From(data)` | 创建Builder实例 | `k.From([]int{1,2,3})` |
| `Build()` | 结束链式,返回结果 | `.Build()` |
| `Sort(less)` | 自定义排序 | `.Sort(func(a,b int)bool{return a<b})` |
| `Filter(pred)` | 过滤 | `.Filter(func(item int, idx int)bool{return item>0})` |
| `Distinct()` | 去重 | `.Distinct()` |
| `DistinctByField(field)` | 按字段去重 | `.DistinctByField("Name")` |
| `ForEach(fn)` | 遍历 | `.ForEach(func(item, idx){...})` |
| `Reverse()` | 反转 | `.Reverse()` |
| `Take(n)` | 取前N个 | `.Take(3)` |
| `TakeRight(n)` | 取后N个 | `.TakeRight(2)` |
| `Skip(n)` | 跳过前N个 | `.Skip(2)` |
| `SkipRight(n)` | 跳过后N个 | `.SkipRight(1)` |
| `Shuffle()` | 随机打乱 | `.Shuffle()` |
| `Pull(values...)` | 移除指定值 | `.Pull(1, 3, 5)` |
| `PullAt(indexes...)` | 移除指定索引 | `.PullAt(0, 2)` |
| `OrderBy(fields, asc)` | 多字段排序 | `.OrderBy([]string{"Age","ID"}, []bool{true,false})` |
| `Map(fn)` | 类型映射 | `.Map(func(item T, idx int) T)` |
| `MapToAny(fn)` | 映射为任意类型 | `.MapToAny(func(item T, idx int)any)` |

---

## 树形结构 (Tree)

位于 `k/tree.go`

通用泛型生成树结构。

### 定义树节点接口

```go
type TreeNodeInterface interface {
    GetID() int64
    GetParentID() int64
    AddChild(TreeNodeInterface)
    SetIsLeaf(bool)
}
```

### 构建树

```go
// nodes: 节点列表, rootId: 根节点ID(可选)
rootNodes := k.BuildTree(nodes)
rootNodes := k.BuildTree(nodes, 0) // 指定根节点ID
```

---

## 时间处理 (Time)

位于 `k/time.go`

### 时间转换

| 函数 | 说明 | 示例 |
|------|------|------|
| `DateToTimesInt10(date)` | 时间转10位时间戳(秒) | `DateToTimesInt10(time.Now())` → 1704067200 |
| `DateToTimesInt13(date)` | 时间转13位时间戳(毫秒) | `DateToTimesInt13(time.Now())` → 1704067200000 |
| `DateStrToTime(dateStr)` | 日期字符串转time | `DateStrToTime("2024-01-01")` → time.Time |
| `DateTimeStrToTime(dateTimeStr)` | 日期时间字符串转time | `DateTimeStrToTime("2024-01-01 12:00:00")` → time.Time |
| `CheckDateTime(timeStr)` | 检查时间格式是否正确 | `CheckDateTime("2024-01-01")` → true |

### 时间计算

| 函数 | 说明 | 示例 |
|------|------|------|
| `DiffBetweenTwoDays(start, end)` | 计算两个日期之间的天数 | `DiffBetweenTwoDays(date1, date2)` → 相差天数 |

### LocalTime 自定义类型

支持JSON序列化/反序列化的自定义时间类型。

```go
type LocalTime struct {
    time.Time
}

// 方法
t.IsZero()          // 是否为空
t.StartOfDay()      // 获取当天开始时间 00:00:00
t.EndOfDay()        // 获取当天结束时间 23:59:59
```

支持解析格式:
- `YYYY-MM-DD`
- `YYYY-MM-DD HH:MM:SS`
- Unix时间戳(10位/13位)
- RFC3339

---

## 随机数 (Random)

位于 `k/random.go`

| 函数 | 说明 | 示例 |
|------|------|------|
| `RandInt(min, max)` | 生成指定范围内的随机整数 | `RandInt(1, 100)` → 随机1-100 |
| `RandString(length)` | 生成指定长度的随机字母 | `RandString(10)` → "aBcDeFgHiJ" |
| `RandUpper(length)` | 生成指定长度的大写字母 | `RandUpper(10)` → "ABCDEFGHIJ" |
| `RandLower(length)` | 生成指定长度的小写字母 | `RandLower(10)` → "abcdefghij" |
| `RandNumeral(length)` | 生成指定长度的数字 | `RandNumeral(6)` → "123456" |
| `RandNumeralOrLetter(length)` | 生成指定长度的字母数字 | `RandNumeralOrLetter(10)` → "a1B2c3D4e5" |
| `GetRandomNum(min, max)` | 生成随机数字(包括边界) | `GetRandomNum(1, 10)` → 1-10之间 |

---

## 重试机制 (Retry)

位于 `k/retry.go`

### 配置选项

| 选项 | 说明 | 默认值 |
|------|------|--------|
| `WithMaxRetries(n)` | 最大重试次数 | 5 |
| `WithRetryDelay(d)` | 基础重试延迟 | 1秒 |
| `WithDelayMultiplier(m)` | 延迟倍数(指数退避) | 1.0 |
| `WithMaxTime(d)` | 最大总执行时间 | 10分钟 |

### 基本用法

```go
err := k.Retry(
    func(args ...any) (any, error) {
        // 执行操作
        resp, err := callAPI()
        if err != nil {
            return nil, err
        }
        return resp, nil
    },
    func(data any) {
        // 成功回调
        fmt.Println("成功:", data)
    },
    k.WithMaxRetries(3),
    k.WithRetryDelay(2*time.Second),
    k.WithDelayMultiplier(2.0),
)
```

### 不可重试错误

```go
// 包装不可重试的错误
return nil, k.NonRetryable(errors.New("认证失败"))

// 判断是否为不可重试错误
if k.IsNonRetryable(err) {
    // 立即终止重试
}
```

### 带Context的重试

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

err := k.RetryWithContext(ctx, operationFn, successFn, options...)
```

---

## HTTP 客户端 (HttpClient)

位于 `k/http_client.go`

生产级HTTP客户端，支持Builder模式构建。

### 快速开始

```go
// 1. 构建客户端
client, err := k.NewClient("https://api.example.com").
    Timeout(10 * time.Second).
    BearerToken(func() string { return tokenStore.Get() }).
    Retry(k.WithMaxRetries(3), k.WithRetryDelay(500*time.Millisecond)).
    CircuitBreaker(k.NewCircuitBreaker()).
    Logger(log.Printf).
    Build()

// 2. 发送请求
resp, err := client.Get("/users",
    k.R().QueryParams(map[string]string{"page": "1"}).ExpectStatus(200),
)

// 3. 读取响应
var users []User
err = client.ReadJSON(resp, &users)
```

### 主要特性

- **链式配置**: 所有配置支持链式调用
- **请求级参数**: 通过 `R()` 构建，与客户端配置分离
- **重试机制**: 内置指数退避重试
- **熔断器**: 防止雪崩效应
- **限速**: 防止请求过于频繁
- **缓存**: 支持响应缓存
- **日志**: 内置日志记录
- **指标**: 内置指标收集

### 请求方法

| 方法 | 说明 |
|------|------|
| `client.Get(path, req)` | GET请求 |
| `client.Post(path, req)` | POST请求 |
| `client.Put(path, req)` | PUT请求 |
| `client.Delete(path, req)` | DELETE请求 |
| `client.Patch(path, req)` | PATCH请求 |
| `client.Head(path, req)` | HEAD请求 |
| `client.Options(path, req)` | OPTIONS请求 |

### 请求配置 (R())

```go
req := k.R().
    QueryParams(map[string]string{"key": "value"}).     // 查询参数
    PathParams(map[string]string{"id": "123"}).       // 路径参数
    Headers(map[string]string{"X-Token": "xxx"}).      // 请求头
    Body(map[string]any{"name": "test"}).              // JSON Body
    FormData(map[string]any{"key": "value"}).           // Form表单
    ExpectStatus(200).                                 // 期望状态码
    Context(ctx).                                      // 上下文
```

---

## 文件目录 (Folder)

位于 `k/folder.go`

| 函数 | 说明 | 示例 |
|------|------|------|
| `PathExists(path)` | 判断文件或目录是否存在 | `PathExists("/tmp/test")` → (true, nil) |
| `CreateDir(path)` | 创建目录(递归) | `CreateDir("/tmp/sub/dir")` |
| `GetCurrentDirectory()` | 获取当前工作目录 | `GetCurrentDirectory()` → "/path/to/cwd" |
| `GetFileSize(path)` | 获取文件大小(字节) | `GetFileSize("/tmp/file.txt")` → 1024 |
| `GetFileExt(filename)` | 获取文件扩展名 | `GetFileExt("test.txt")` → ".txt" |
| `GetFileName(filename)` | 获取文件名(不含扩展名) | `GetFileName("test.txt")` → "test" |
| `GetFileNameWithExt(path)` | 获取文件名(含扩展名) | `GetFileNameWithExt("/tmp/test.txt")` → "test.txt" |

---

## Excel 导出 (Excel)

位于 `k/excel.go`

### 定义列

```go
type User struct {
    Name   string
    Age    int
    Salary float64
}

columns := []k.DefExcelColumn[User]{
    {Header: "姓名", GetValue: func(item User, index int) interface{} { return item.Name }},
    {Header: "年龄", GetValue: func(item User, index int) interface{} { return item.Age }},
    {Header: "薪资", GetValue: func(item User, index int) interface{} { return item.Salary }, Sum: true, NumberFormat: true, Decimal: 2},
}
```

### 列配置选项

| 字段 | 说明 |
|------|------|
| `Header` | 表头名称 |
| `Hide` | 是否隐藏列 |
| `Width` | 列宽(不设置则自动计算) |
| `GetValue` | 获取单元格数据 |
| `HeaderStyle` | 表头样式 |
| `CellStyle` | 单元格样式 |
| `MergeRow` | 向下合并行数 |
| `MergeCol` | 向右合并列数 |
| `NumberFormat` | 是否数字格式化 |
| `Decimal` | 小数位数 |
| `Sum` | 是否在底部求和 |

### 导出方式

```go
// 导出到Excel文件
f, err := k.ExportToExcel[User]("用户列表", true, columns, users)
f.SaveAs("users.xlsx")

// 直接通过HTTP响应
k.ExportExcelToHttp[User](w, "用户列表", true, columns, users)
```

---

## 验证码 (Captcha)

位于 `k/captcha/captcha.go`

### 设置存储

```go
import "github.com/kuangshp/go-utils/k/captcha"

// 使用缓存存储
cache := NewCacheStore(redisAdapter, 5*time.Minute)
captcha.SetStore(cache)
```

### 生成验证码

```go
// 生成字母验证码
id, b64s, answer, err := captcha.DriverStringFunc()

// 生成数字验证码
id, b64s, answer, err := captcha.DriverDigitFunc()
```

### 验证验证码

```go
// 验证(clear=true验证后删除)
ok := captcha.Verify(id, code, true)
```

---

## 图片缓存存储 (Store)

位于 `k/store/store.go`

用于base64Captcha的缓存适配器。

### 定义缓存接口

```go
type AdapterCache interface {
    Set(key string, value string, expiration time.Duration) error
    Get(key string) (string, error)
    Del(key string) error
}
```

### 创建存储

```go
store := store.NewCacheStore(redisAdapter, 5*time.Minute)
```

### 内置实现

| 实现 | 文件 | 说明 |
|------|------|------|
| `MemoryStore` | `k/store/memory.go` | 内存缓存 |
| `TypeStore` | `k/store/type.go` | 类型化存储 |

---

## 身份证操作 (Card)

位于 `k/card.go`

| 函数 | 说明 | 示例 |
|------|------|------|
| `GetBirthFromIDCard(idCard)` | 从身份证号获取出生日期 | `GetBirthFromIDCard("110101199001011234")` → "1990-01-01" |
| `GetAgeFromIDCard(idCard)` | 从身份证号获取年龄 | `GetAgeFromIDCard("110101199001011234")` → 34 |
| `GetGenderFromIDCard(idCard)` | 从身份证号获取性别 | `GetGenderFromIDCard("110101199001011234")` → "男"或"女" |

---

## IP 地址操作 (IP)

位于 `k/ip.go`

### 获取客户端IP

```go
ip := k.ClientIP(r) // *http.Request
```

支持从以下Header获取:
- `X-Forwarded-For`
- `X-Real-IP`
- `RemoteAddr`

### IP地址归属地查询

```go
province, city, detail := k.GetIpToAddress("8.8.8.8")
// province: 省
// city: 市
// detail: IP详细信息(包含国家、ISP、经纬度等)
```

---

## URL 操作 (URL)

位于 `k/url.go`

| 函数 | 说明 | 示例 |
|------|------|------|
| `EncodeURIComponent(str)` | URL编码 | `EncodeURIComponent("你好")` → "%E4%BD%A0%E5%A5%BD" |
| `URLToMap(queryStr)` | URL查询字符串转map | `URLToMap("name=hello&age=20")` → map[string]string{"name":"hello","age":"20"} |

---

## 条件工具 (If)

位于 `k/if.go`

| 函数 | 说明 | 示例 |
|------|------|------|
| `If(condition, trueVal, falseVal)` | 条件选择 | `If(a > b, a, b)` → 较大值 |
| `IfLazy(condition, trueFunc, falseValue)` | 惰性求值条件选择 | `IfLazy(cond, func() int { return heavyCalc() }, 0)` |

---

## 加密 (Encryption)

位于 `k/encryption.go`

### AES加解密

```go
// 加密
origData := []byte("需要加密的内容")
key := []byte("16位密钥字符!") // 16/24/32位分别对应AES-128/192/256
encrypted, err := k.AesEcrypt(origData, key)

// 解密
decrypted, err := k.AesDeCrypt(encrypted, key)
```

### 填充模式

| 函数 | 说明 |
|------|------|
| `PKCS7Padding(data, blockSize)` | PKCS7填充 |
| `PKC7UnPadding(data)` | 移除填充 |

---

## 使用示例

### 完整示例 - 切片操作

```go
package main

import (
    "fmt"
    "github.com/kuangshp/go-utils/k"
)

type User struct {
    Name string
    Age  int
    Dept string
}

func main() {
    users := []User{
        {"张三", 25, "技术部"},
        {"李四", 30, "技术部"},
        {"王五", 28, "市场部"},
        {"赵六", 35, "技术部"},
    }

    // 过滤技术部员工并按年龄排序
    result := k.From(users).
        Filter(func(item User, idx int) bool { return item.Dept == "技术部" }).
        Sort(func(a, b User) bool { return a.Age < b.Age }).
        Build()

    fmt.Println(result)
}
```

### 完整示例 - HTTP客户端

```go
package main

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "time"

    "github.com/kuangshp/go-utils/k"
)

type User struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
}

func main() {
    client, err := k.NewClient("https://api.example.com").
        Timeout(10 * time.Second).
        Retry(k.WithMaxRetries(3), k.WithRetryDelay(500*time.Millisecond)).
        Logger(log.Printf).
        Build()

    if err != nil {
        panic(err)
    }

    resp, err := client.Get("/users",
        k.R().QueryParams(map[string]string{"page": "1", "size": "10"}).ExpectStatus(200),
    )
    if err != nil {
        panic(err)
    }

    var users []User
    if err := client.ReadJSON(resp, &users); err != nil {
        panic(err)
    }

    fmt.Printf("获取到 %d 个用户\n", len(users))
}
```

