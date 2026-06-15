package monitor

import (
	"ecloud_computer_auto_boot/pkg/ecloud"
	"ecloud_computer_auto_boot/pkg/service/logger"
	"ecloud_computer_auto_boot/pkg/util"
	"fmt"

	"gitlab.ecloud.com/ecloud/ecloudsdkcomputer/model"
)

// check 执行监控检查
func (t *MonitorTask) check() {
	t.LastCheck = util.Log().Info("[监控检查] 账号 %s 开始检查", t.AccountName)

	if t.AccountType == "public" {
		t.checkPublic()
	} else {
		t.checkAPI()
	}
}

// checkPublic 公众版监控
func (t *MonitorTask) checkPublic() {
	if t.PublicClient == nil {
		t.Status = "error"
		t.LastEvent = "客户端未初始化"
		return
	}

	resp, err := t.PublicClient.GetDeviceInfo()
	if err != nil {
		util.Log().Error("[%s] 获取设备信息失败: %s", t.AccountID, err)
		t.Status = "error"
		t.LastEvent = fmt.Sprintf("获取设备信息失败: %s", err.Error())
		return
	}

	if !resp.Success() {
		util.Log().Error("[%s] 获取云电脑列表失败: %s", t.AccountID, resp.ErrorMessage)
		t.Status = "error"
		t.LastEvent = resp.ErrorMessage
		return
	}

	monitorAll := len(t.MachineIDs) == 0
	body := resp.GetBody()
	machineList := body["machineList"].([]interface{})

	for _, machine := range machineList {
		info := machine.(map[string]interface)

		computer := ecloud.ComputerInfo{
			MachineID:   info["machineId"].(string),
			MachineName: info["machineName"].(string),
			CompanyCode: info["companyCode"].(string),
			Status:      ecloud.GetComputerStatus(info["resourceStatus"].(string)),
		}

		util.Log().Debug("[%s] 机器: %s, 状态: %s", t.AccountID, computer.MachineName, computer.Status)

		// 检查是否需要监控此机器
		if !monitorAll && !contains(t.MachineIDs, computer.MachineID) {
			continue
		}

		// 如果机器已关机，尝试开机
		if computer.Status == ecloud.ResourceStatusShutdown {
			util.Log().Info("[%s] 检测到机器 %s 已关机，请求开机", t.AccountID, computer.MachineName)

			t.logManager.Log(logger.Event{
				Type:      logger.EventBoot,
				AccountID: t.AccountID,
				Message:   fmt.Sprintf("检测到机器 %s 已关机，尝试开机", computer.MachineName),
				Details: map[string]interface{}{
					"machine_id":   computer.MachineID,
					"machine_name": computer.MachineName,
				},
				Status: "info",
			})

			_, err := t.PublicClient.OperateComputer(computer, ecloud.ComputerOperationAvailable)
			if err != nil {
				util.Log().Error("[%s] 机器 %s 开机失败: %s", t.AccountID, computer.MachineName, err)
				t.LastEvent = fmt.Sprintf("开机失败: %s", computer.MachineName)

				t.logManager.Log(logger.Event{
					Type:      logger.EventBootFailed,
					AccountID: t.AccountID,
					Message:   fmt.Sprintf("机器 %s 开机失败: %s", computer.MachineName, err.Error()),
					Details: map[string]interface{}{
						"machine_id":   computer.MachineID,
						"machine_name": computer.MachineName,
						"error":        err.Error(),
					},
					Status: "failed",
				})
			} else {
				util.Log().Info("[%s] 机器 %s 已完成开机操作", t.AccountID, computer.MachineName)
				t.LastEvent = fmt.Sprintf("开机成功: %s", computer.MachineName)

				t.logManager.Log(logger.Event{
					Type:      logger.EventBootSuccess,
					AccountID: t.AccountID,
					Message:   fmt.Sprintf("机器 %s 开机成功", computer.MachineName),
					Details: map[string]interface{}{
						"machine_id":   computer.MachineID,
						"machine_name": computer.MachineName,
					},
					Status: "success",
				})
			}
		}
	}

	t.Status = "running"
}

// checkAPI 政企版监控
func (t *MonitorTask) checkAPI() {
	if t.APIClient == nil {
		t.Status = "error"
		t.LastEvent = "客户端未初始化"
		return
	}

	page := int32(1)
	failedCnt := 0

	for {
		if failedCnt >= 3 {
			t.Status = "error"
			t.LastEvent = "请求失败次数过多"
			return
		}

		response, err := t.findMachineOnOpenAPI(page)
		if err != nil {
			failedCnt++
			util.Log().Error("[%s] 请求失败: %s", t.AccountID, err)
			continue
		}

		if *response.ErrorCode != "" {
			failedCnt++
			util.Log().Error("[%s] 请求失败: %s", t.AccountID, *response.ErrorMessage)
			continue
		}

		respData := *response.Body.Data
		for _, instance := range respData {
			machineId := *instance.MachineId
			machineName := *instance.MachineName

			// 检查是否需要监控
			if len(t.MachineIDs) > 0 && !contains(t.MachineIDs, machineId) {
				continue
			}

			util.Log().Debug("[%s] 机器: %s, 状态: %s", t.AccountID, machineName, *instance.MachineStatus)

			if *instance.MachineStatus == "shutdown" {
				util.Log().Info("[%s] 检测到机器 %s 已关机，请求启动", t.AccountID, machineName)

				t.logManager.Log(logger.Event{
					Type:      logger.EventBoot,
					AccountID: t.AccountID,
					Message:   fmt.Sprintf("检测到机器 %s 已关机，尝试开机", machineName),
					Details: map[string]interface{}{
						"machine_id":   machineId,
						"machine_name": machineName,
					},
					Status: "info",
				})

				resp, err := t.startupMachineOnOpenAPI(machineId)
				if err != nil {
					util.Log().Error("[%s] 机器 %s 启动失败: %s", t.AccountID, machineName, err)
					t.LastEvent = fmt.Sprintf("开机失败: %s", machineName)

					t.logManager.Log(logger.Event{
						Type:      logger.EventBootFailed,
						AccountID: t.AccountID,
						Message:   fmt.Sprintf("机器 %s 启动失败: %s", machineName, err.Error()),
						Details: map[string]interface{}{
							"machine_id":   machineId,
							"machine_name": machineName,
							"error":        err.Error(),
						},
						Status: "failed",
					})
					continue
				}

				if *resp.ErrorCode != "" {
					util.Log().Error("[%s] 机器 %s 启动失败: %s", t.AccountID, machineName, *resp.ErrorMessage)
					t.LastEvent = fmt.Sprintf("开机失败: %s", machineName)

					t.logManager.Log(logger.Event{
						Type:      logger.EventBootFailed,
						AccountID: t.AccountID,
						Message:   fmt.Sprintf("机器 %s 启动失败: %s", machineName, *resp.ErrorMessage),
						Details: map[string]interface{}{
							"machine_id":   machineId,
							"machine_name": machineName,
							"error":        *resp.ErrorMessage,
						},
						Status: "failed",
					})
				} else {
					util.Log().Info("[%s] 机器 %s 已完成开机操作", t.AccountID, machineName)
					t.LastEvent = fmt.Sprintf("开机成功: %s", machineName)

					t.logManager.Log(logger.Event{
						Type:      logger.EventBootSuccess,
						AccountID: t.AccountID,
						Message:   fmt.Sprintf("机器 %s 开机成功", machineName),
						Details: map[string]interface{}{
							"machine_id":   machineId,
							"machine_name": machineName,
						},
						Status: "success",
					})
				}
			}
		}

		// 页尾判定
		if *response.Body.TotalSize <= page*50 {
			break
		}

		page++
	}

	t.Status = "running"
}

// findMachineOnOpenAPI 查询机器列表（政企版）
func (t *MonitorTask) findMachineOnOpenAPI(page int32) (*model.GetResourceListResponse, error) {
	body := &model.GetResourceListBody{}
	body.SetPage(page).SetPageSize(50)

	request := &model.GetResourceListRequest{
		GetResourceListBody: body,
	}

	return t.APIClient.GetResourceList(request)
}

// startupMachineOnOpenAPI 启动机器（政企版）
func (t *MonitorTask) startupMachineOnOpenAPI(machineId string) (*model.OperateMachineByAvailableResponse, error) {
	request := &model.OperateMachineByAvailableRequest{}
	request.OperateMachineByAvailableQuery.SetMachineId(machineId)

	return t.APIClient.OperateMachineByAvailable(request)
}

// contains 检查字符串切片是否包含元素
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
