package writer

import (
	"nimo-shake/common"

	LOG "github.com/vinllen/log4go"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

type Writer interface{
	// create table
	CreateTable(tableDescribe *dynamodb.TableDescription) error
	// drop table
	DropTable() error
	// write bulk data
	WriteBulk(input []interface{}) error
	// insert
	Insert(input []interface{}, index []interface{}) error
	// delete
	Delete(input []interface{}) error
	// update
	Update(input []interface{}, index []interface{}) error
	// close
	Close()
}

func NewWriter(name, address string, ns utils.NS, logLevel string) Writer {
	switch name {
	case utils.TargetTypeMongo:
		return NewMongoWriter(name, address, ns)
	case utils.TargetTypeAliyunDynamoProxy:
		return NewDynamoProxyWriter(name, address, ns, logLevel)
	default:
		LOG.Crashf("unknown writer[%v]", name)
	}
	return nil
}