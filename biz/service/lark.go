package service

import (
	"context"
	"fmt"

	"github.com/cloudwego/hertz/pkg/common/json"
	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"github.com/zhikongming/stock/biz/config"
	"github.com/zhikongming/stock/biz/model"
)

func SendLarkMessage(ctx context.Context, message *model.LarkMessage) error {
	header := map[string]string{
		"Content-Type": "application/json",
	}
	larkConfig := config.GetLarkConfig()
	if larkConfig == nil {
		return fmt.Errorf("lark config is nil")
	}
	_, err := DoPost(ctx, larkConfig.GroupRobotURL, nil, header, message)
	if err != nil {
		return err
	}
	return nil
}

func SendTestLarkMessage(ctx context.Context, message *model.LarkMessage) error {
	larkConfig := config.GetLarkConfig()
	if larkConfig == nil {
		fmt.Printf("lark config: %v\n", larkConfig)
		return fmt.Errorf("lark config is nil")
	}

	// 创建客户端
	client := lark.NewClient(larkConfig.AppID, larkConfig.AppSecret)

	// 构建文本消息
	// 注意：接收者ID需填写用户的 open_id 或 user_id
	receiveIdType := "open_id"
	receiveId := larkConfig.TestReceiveID // 替换为实际的接收者ID

	contentBytes, _ := json.Marshal(message.Card)
	content := string(contentBytes)

	req := larkim.NewCreateMessageReqBuilder().
		ReceiveIdType(receiveIdType).
		Body(larkim.NewCreateMessageReqBodyBuilder().
			ReceiveId(receiveId).
			MsgType(larkim.MsgTypeInteractive).
			Content(content).
			Build()).
		Build()

	// 4. 发送消息
	resp, err := client.Im.Message.Create(context.Background(), req)
	if err != nil {
		fmt.Printf("create message error: %v\n", err)
		return err
	}

	if !resp.Success() {
		fmt.Printf("create message error: %v\n", resp)
		return fmt.Errorf("业务错误，code: %d, msg: %s", resp.Code, resp.Msg)
	}
	return nil
}
