package service

import (
	"encoding/csv"
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"redis_key_analysis/internal/conf"
	"strings"
	"time"
)

// 结果集的结构体
type Field struct {
	Key     string // map key
	ColName string // CSV 列名
}

type CSVResult struct {
	Data     []map[string]any
	BasePath string
	FileName string
	FullPath string
	Kind     string
}

func NewCSVResult(data []map[string]any, fileName, kind string) CSVResult {
	appConf := conf.GetAppConfig()
	if appConf.ExportFilePath == "" {
		appConf.ExportFilePath = "/tmp"
	}
	return CSVResult{
		Data:     data,
		BasePath: appConf.ExportFilePath,
		FileName: fileName + "_" + kind,
		Kind:     kind,
	}
}

func (cr *CSVResult) prefixFields() []Field {
	// [object Object]	类型	占有内存	Key数量	元素数量
	return []Field{
		{"Prefix", "[object Object]"},
		{"Type", "类型"},
		{"Bytes", "占有内存"},
		{"KeyNum", "Key数量"},
		{"Count", "元素数量"},
	}
}

func (cr *CSVResult) bigKeysFields() []Field {
	// [object Object]	类型	占有内存	Key数量	元素数量
	return []Field{
		{"Key", "Key"},
		{"NodeId", "节点ID"},
		{"Type", "类型"},
		{"Encoding", "Encoding"},
		{"Bytes", "占有内存"},
		{"Count", "元素数量"},
		{"MaxLength", "最大元素的长度"},
		{"ExpirationTimeMillis", "过期时间"},
		{"Db", "DB"},
	}
}

func (cr *CSVResult) topPrefixColNames() []string {
	// [object Object]	类型	占有内存	Key数量	元素数量
	fields := cr.prefixFields()
	colNames := make([]string, len(fields))
	for k, f := range fields {
		colNames[k] = f.ColName
	}
	return colNames
}

func (cr *CSVResult) topBigMemColNames() []string {
	// [object Object]	类型	占有内存	Key数量	元素数量
	fields := cr.bigKeysFields()
	colNames := make([]string, len(fields))
	for k, f := range fields {
		colNames[k] = f.ColName
	}
	return colNames
}

// 提取行数据成切片(当前行)
func generateRowsData(record map[string]any, headers []Field) []string {
	row := make([]string, 0, len(headers))
	for _, col := range headers {
		var rowData string
		// TODO: 格式化
		switch col.Key {
		case "Bytes":
			rowData = bytesToHumanReadable(record[col.Key])
		case "Count":
			rowData = countToHumanReadable(record[col.Key])
		default:
			rowData = fmt.Sprintf("%v", record[col.Key])
		}
		row = append(row, rowData)
	}
	return row
}

// 判断路径是否存在
func pathIsExist(base string) error {
	// 创建文件，不存在目录则创建
	_, err := os.Stat(base)
	if err != nil {
		if os.IsNotExist(err) {
			err := os.MkdirAll(base, 0755)
			if err != nil {
				return errors.New("create file path failed, " + err.Error())
			}
		} else {
			return errors.New("unknown error, " + err.Error())
		}
	}
	return nil
}

// 转换高可读性的显示
func countToHumanReadable(count any) string {
	assertVal, ok := count.(float64)
	if !ok {
		return "?"
	}
	countFloat64 := assertVal
	if countFloat64 == 0 {
		return "0"
	}

	units := []string{"", "K", "M", "B", "T"} // 千、百万、十亿、万亿
	magnitude := math.Floor(math.Log10(math.Abs(countFloat64)) / 3)

	if magnitude >= float64(len(units)) {
		magnitude = float64(len(units) - 1)
	}
	if magnitude < 0 {
		magnitude = 0
	}

	unitIndex := int(magnitude)
	scaled := countFloat64 / math.Pow(10, magnitude*3)

	// 如果是整数，不显示小数；否则保留两位
	if scaled == float64(int64(scaled)) {
		return fmt.Sprintf("%.0f%s", scaled, units[unitIndex])
	}
	return fmt.Sprintf("%.2f%s", scaled, units[unitIndex])
}

func bytesToHumanReadable(bytes any) string {
	assertVal, ok := bytes.(float64)
	if !ok {
		return "?"
	}
	bytesfloat64 := assertVal
	if bytesfloat64 == 0 {
		return "0 B"
	}

	const unit = 1024
	sizes := []string{"B", "KB", "MB", "GB", "TB", "PB", "EB"}

	// 计算单位级别（log_1024(bytes)）
	exponent := int(math.Floor(math.Log(bytesfloat64) / math.Log(unit)))
	if exponent > len(sizes)-1 {
		exponent = len(sizes) - 1
	}

	// 转换为对应单位的数值
	value := bytesfloat64 / math.Pow(unit, float64(exponent))

	// 格式化：保留 2 位小数，但整数不显示小数
	if value == float64(int64(value)) {
		return fmt.Sprintf("%.0f %s", value, sizes[exponent])
	}
	return fmt.Sprintf("%.2f %s", value, sizes[exponent])
}

// 转换成CSV文件并存储在本地
func (cr *CSVResult) Convert() error {
	err := pathIsExist(cr.BasePath)
	if err != nil {
		return err
	}

	if cr.BasePath == "" {
		appConf := conf.GetAppConfig()
		cr.BasePath = appConf.ExportFilePath
	}
	if beforePath, ok := strings.CutSuffix(cr.BasePath, "/"); ok {
		cr.BasePath = beforePath
	}
	now := time.Now().Format("200601021504")
	if cr.FileName == "" {
		cr.FileName = "unknown_redis_analysis_result_" + now
	}
	cr.FileName = cr.FileName + "_" + now + ".csv" // 完整文件名
	absFilePath := cr.BasePath + "/" + cr.FileName // 绝对路径
	cr.FullPath = absFilePath
	f, err := os.Create(absFilePath)
	if err != nil {
		return err
	}
	defer f.Close()

	// 避免Window Excel打开中文乱码
	f.WriteString("\xEF\xBB\xBF")

	// 制作表头数据
	w := csv.NewWriter(f)
	defer w.Flush()
	if cr.Data == nil {
		// 空数据直接返回
		return nil
	}

	var colNames []string
	var colKeys []Field
	switch strings.ToLower(cr.Kind) {
	case "topprefix":
		colNames = cr.topPrefixColNames()
		colKeys = cr.prefixFields()
	case "topbigmem":
		colNames = cr.topBigMemColNames()
		colKeys = cr.bigKeysFields()
	default:
		return errors.New("unknown result kind")
	}

	// 写入表头
	if err := w.Write(colNames); err != nil {
		log.Println("write headers csv file is error,", err.Error())
		return err
	}
	// 写入结果集数据
	for _, row := range cr.Data {
		rowData := generateRowsData(row, colKeys)
		err := w.Write(rowData)
		if err != nil {
			log.Println("write row data csv file is error,", err.Error())
			return errors.New("write data failed")
		}
	}
	return nil
}
