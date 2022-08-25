## 数据类型

### 基本类型

#### 不带默认值
Int  整数 (-1, 0, 1, 200 ...)
Float  浮点数 (-0.5, 0.1, 0.2, 10.05 ...)
Bool  布尔值 (1, 0)
Str 字符串  (hello, Scale, NpcId, ...)

#### 带默认值
Int=123  整数 (-1, 0, 1, 200 ...)
Float=10.123  浮点数 (-0.5, 0.1, 0.2, 10.05 ...)
Bool=1  布尔值 (1, 0)
Str=whosyourdaddy 字符串  (hello, Scale, NpcId, ...)

### 进阶类型

####List开头, 数据以英文","或者英文";"分割。  (分号";"分割是为了兼容老数据，建议使用逗号","")

##### 简单
+ List(Int)  整数列表   如1,2,3  如4;5;6
+ List(Bool) 布尔值列表 如1,0,0  如0;0;1
+ List(Float)浮点数列表 如1.1,2.1,3.1 如2.2;3.2;4.2
+ List(Str)  字符串列表 如man,woman 如one;two;three

##### 进阶
+ List(List(Int))   如 [1,2,3],[4,5]  如[7,8,9]  如 [10],[13,14,15],[12,18]
+ List(List(Bool))  如 [1],[0,1]  如[1,1,0]  如 [1],[1,1,1],[0,0]
+ List(List(Float)) 如 [1.1,2,3.2],[4,5]  如[7.9,0.8,9]  如 [1.0],[1.3,1.4,1.5],[12,18]
+ List(List(Str))   如 [man,woman],[one;two;three] 如 [one;two;three]
+ List(List(List(List(List...)))) 任意数量嵌套 如 [ [[1,2],[3,4]], [[5,6],[7,8]] ]

#####Dict开头, 每个key-value对以英文","分割

##### 简单
+ Dict(a:Int, b:Float, c:Bool, d:Str) 如 a=1, b=1.2, c=1, d=nihao
+ Dict(a:Int, b:Int, c:Str, d:Str) 如 a=10, b=10, c=up, d=down

##### 进阶
+ Dict(a:List(Int))         如 a=7,8,9
+ Dict(a:Str, b:List(Int))  如 a=pos, b=9,10
+ Dict(a:List(Str), b:List(List(Int)))  如 a=step,speed, b=[1,2],[3,4,5],[6]
+ Dict(a:List(List(List...)))  任意List嵌套在内

#####Enum开头

Enum是一个枚举类型，是一种数据罗列方式。
比如类型： Enum(None:0, Apple:1, Banana:2, Orange:3)
    数据：Apple 导出是1  Orange 导出是3
**不填的单元格，默认返回0**，和上面类型中填None一个效果。


#####Func开头

Func是一个函数类型，会直接导出成函数使用。
数据填满足lua语法的表达式即可，一般意义上的数学表达式都满足lua语法。




#### 进阶类型总结
List可以作为子节点，Dict不可以作为子节点。


## Excel配表特性

### skiprow
一般作为表头注释等作用，不纳入任何逻辑中，导表工具完全忽略的行。
可以有多行skiprow。

### 类型行
skiprow结束后的下一行是类型行，按照上述**数据类型**进行配置。

### 空行
类型行下面一行是空行，目前无作用，兼容历史遗留数据。

### 程序字段行
紧接着是程序使用的字段名，命名必须是字母开头，字母+数字的组合，尽量不要使用特殊字符。
**程序字段不填，这一列不导表。**
特殊字段**ExportTable**, 必须是Bool类型，用来标记对应单行数据是否导表。 某些表在测试阶段需要控制每一行是否导表，可以加该字段，类型是Bool=1，这样填0的行不导表。 没有ExportTable字段的表默认每一行都导表，没有需求时完全不用关注这个字段。

### 数据
程序字段行下面就是数据内容了。
+ 数据按照**类型行**做解析
+ 数据行可以是**空行**，不影响数据解析，可以用来当做数据分析人员的分块查看作用



## 特殊字段劫持 (程序关注)

目前有需求，在导表阶段就要处理一些简单数据。
这种时候，已经不是类型转换的问题，可能有一些逻辑计算等情况，导表引擎判断不了。

可以在hook目录添加 [数据名].lua 文件, 代表[数据名]的表有字段要自己特殊处理。
比如  AppleData.lua  表示AppleData数据，有字段由lua逻辑来处理，不用导表引擎处理。
比如  AppleData有个字段Position需要经过计算再导出，函数签名要求是AppleData_Position(string)
```
-- text是字段Position的配表内容, 比如 1#2,3#4 
function AppleData_Position(text)
    local r = {elements = {}}
    local arr = split(text, ",")
    for _, pairStr in ipairs(arr) do
        local x, y = unpack(split(pairStr))
        r.elements[#r.elements+1] = 10000 * x + y
    return r
end
```

## 导出数据特性 (程序关注)

所有导出的表字段不会是nil，无需再进行判断。 比如List为空就是一个空列表。
