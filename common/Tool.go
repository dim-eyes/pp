package common

import (
	"reflect"
	"strconv"
	"time"
	"unsafe"
)

func StructToMap(value interface{}) map[string]interface{} {
	var structMap = make(map[string]interface{})
	typeObj := reflect.TypeOf(value)
	valueObj := reflect.ValueOf(value)

	for i := 0; i < typeObj.NumField(); i++ {
		structMap[typeObj.Field(i).Name] = valueObj.Field(i).Interface()
	}
	return structMap
}

func IsSameDay(timestamp1, timestamp2 int64) bool {
	time1 := time.Unix(timestamp1, 0).Format("20060102")
	time2 := time.Unix(timestamp2, 0).Format("20060102")
	return time1 == time2

}

// BeginOfDay beginning of day
func BeginOfDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

// BeginOfYesterday begin yesterday
func BeginOfYesterday() time.Time {
	return BeginOfDay(Yesterday())
}

// Yesterday 昨天的时间
func Yesterday() time.Time {
	return time.Now().AddDate(0, 0, -1)
}

// StrToTimestamp 时间转时间戳
func StrToTimestamp(timeStr string) (int64, bool) {
	t1 := "2006-01-02 15:04:05"
	stamp, err := time.ParseInLocation(t1, timeStr, time.Local) // 使用parseInLocation将字符串格式化返回本地时区时间
	if err != nil {
		return 0, false
	}
	return stamp.Unix(), true
}

func Str2bytes(s string) []byte {
	x := (*[2]uintptr)(unsafe.Pointer(&s)) // 获取s的起始地址开始后的两个 uintptr 指针
	h := [3]uintptr{x[0], x[1], x[1]}      // 构造三个指针数组
	return *(*[]byte)(unsafe.Pointer(&h))
}

// Week 获取时间的周数 46 47 48
func Week(t time.Time) int {
	_, week := t.ISOWeek()
	return week
}

// DayStr 获取时间的日期 string 示例: 20211218
func DayStr(t time.Time) string {
	return t.Format("20060102")
}

// YesterdayStr 昨天的日期string 示例: 20211218
func YesterdayStr() string {
	return DayStr(Yesterday())
}

// WeekDay 获取时间是周几
// 周一:1 周二:2 周三:3 周四:4 周五:5 周六:6 周日:7
func WeekDay(t time.Time) int {
	week := int(t.Weekday())
	if week == 0 {
		return 7
	}
	return week
}

// InterfaceToString interface 转string
func InterfaceToString(iface interface{}) string {
	var result = ""
	switch v := iface.(type) {
	case string:
		result = v
	case int:
		result = strconv.Itoa(v)
	case float64:
		result = strconv.Itoa(int(v))
	case nil:
		result = ""
	default:
		result = ""
	}
	return result
}

// InterfaceToInt8 interface 转int8
func InterfaceToInt8(iface interface{}) int8 {
	var result = 0
	var err error
	switch v := iface.(type) {
	case string:
		result, err = strconv.Atoi(v)
		if err != nil {
			return 0
		}
	case int:
		result = v
	case float64:
		result = int(v)
	case nil:
		result = 0
	default:
		result = 0
	}
	return int8(result)
}

// InterfaceToInt interface 转int
func InterfaceToInt(iface interface{}) int {
	var result = 0
	var err error
	switch v := iface.(type) {
	case string:
		result, err = strconv.Atoi(v)
		if err != nil {
			return 0
		}
	case int:
		result = v
	case float64:
		result = int(v)
	case nil:
		result = 0
	default:
		result = 0
	}
	return result
}

// InterfaceToInt64 interface 转int
func InterfaceToInt64(iface interface{}) int64 {
	var result int64 = 0
	var err error
	switch v := iface.(type) {
	case string:
		result, err = strconv.ParseInt(v, 10, 64)
		if err != nil {
			return 0
		}
	case int:
		result = int64(v)
	case float64:
		result = int64(v)
	default:
		result = 0
	}
	return result
}

func GetPageOffset(pageNum, pageSize int64) int64 {
	return (pageNum - 1) * pageSize
}
