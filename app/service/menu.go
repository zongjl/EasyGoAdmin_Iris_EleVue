// +----------------------------------------------------------------------
// | EasyGoAdmin敏捷开发框架 [ 赋能开发者，助力企业发展 ]
// +----------------------------------------------------------------------
// | 版权所有 2019~2022 深圳EasyGoAdmin研发中心
// +----------------------------------------------------------------------
// | Licensed LGPL-3.0 EasyGoAdmin并不是自由软件，未经许可禁止去掉相关版权
// +----------------------------------------------------------------------
// | 官方网站: http://www.easygoadmin.vip
// +----------------------------------------------------------------------
// | Author: @半城风雨 团队荣誉出品
// +----------------------------------------------------------------------
// | 版权和免责声明:
// | 本团队对该软件框架产品拥有知识产权（包括但不限于商标权、专利权、著作权、商业秘密等）
// | 均受到相关法律法规的保护，任何个人、组织和单位不得在未经本团队书面授权的情况下对所授权
// | 软件框架产品本身申请相关的知识产权，禁止用于任何违法、侵害他人合法权益等恶意的行为，禁
// | 止用于任何违反我国法律法规的一切项目研发，任何个人、组织和单位用于项目研发而产生的任何
// | 意外、疏忽、合约毁坏、诽谤、版权或知识产权侵犯及其造成的损失 (包括但不限于直接、间接、
// | 附带或衍生的损失等)，本团队不承担任何法律责任，本软件框架禁止任何单位和个人、组织用于
// | 任何违法、侵害他人合法利益等恶意的行为，如有发现违规、违法的犯罪行为，本团队将无条件配
// | 合公安机关调查取证同时保留一切以法律手段起诉的权利，本软件框架只能用于公司和个人内部的
// | 法律所允许的合法合规的软件产品研发，详细声明内容请阅读《框架免责声明》附件；
// +----------------------------------------------------------------------

package service

import (
	"easygoadmin/app/dto"
	"easygoadmin/app/model"
	"easygoadmin/app/vo"
	"easygoadmin/utils"
	"easygoadmin/utils/gconv"
	"easygoadmin/utils/gstr"
	"errors"
	"strconv"
	"strings"
	"time"
)

var Menu = new(menuService)

type menuService struct{}

func (s *menuService) GetList(req dto.MenuQueryReq) ([]model.Menu, error) {
	// 创建查询实例
	query := utils.XormDb.Where("mark=1")
	// 部门名称
	if req.Title != "" {
		query = query.Where("title like ?", "%"+req.Title+"%")
	}
	// 排序
	query = query.OrderBy("sort")
	// 查询列表
	var list []model.Menu
	err := query.Find(&list)
	return list, err
}

func (s *menuService) Add(req dto.MenuAddReq, userId int) (int64, error) {
	if utils.AppDebug() {
		return 0, errors.New("演示环境，暂无权限操作")
	}
	// 实例化对象
	var entity model.Menu
	entity.ParentId = req.ParentId
	entity.Title = req.Title
	entity.Icon = req.Icon
	entity.Path = req.Path
	entity.Component = req.Component
	if req.Target == 2 {
		entity.Target = "_blank"
	} else {
		entity.Target = "_self"
	}
	entity.Permission = req.Permission
	entity.Type = req.Type
	entity.Status = req.Status
	entity.Hide = req.Hide
	entity.Note = req.Note
	entity.Sort = req.Sort
	entity.CreateUser = userId
	entity.CreateTime = time.Now().Unix()
	entity.UpdateUser = userId
	entity.UpdateTime = time.Now().Unix()
	entity.Mark = 1
	// 插入数据
	rows, err := entity.Insert()
	if err != nil || rows == 0 {
		return 0, errors.New("添加失败")
	}
	// 添加节点
	setPermission(req.Type, req.CheckedList, req.Title, req.Path, gconv.Int(entity.Id), userId)
	return rows, nil
}

func (s *menuService) Update(req dto.MenuUpdateReq, userId int) (int64, error) {
	if utils.AppDebug() {
		return 0, errors.New("演示环境，暂无权限操作")
	}
	// 查询记录
	entity := &model.Menu{Id: req.Id}
	has, err := entity.Get()
	if err != nil || !has {
		return 0, err
	}
	entity.ParentId = req.ParentId
	entity.Title = req.Title
	entity.Icon = req.Icon
	entity.Path = req.Path
	entity.Component = req.Component
	if req.Target == 2 {
		entity.Target = "_blank"
	} else {
		entity.Target = "_self"
	}
	entity.Permission = req.Permission
	entity.Type = req.Type
	entity.Status = req.Status
	entity.Hide = req.Hide
	entity.Note = req.Note
	entity.Sort = req.Sort
	entity.UpdateUser = userId
	entity.UpdateTime = time.Now().Unix()
	// 更新数据
	rows, err := entity.Update()
	if err != nil || rows == 0 {
		return 0, errors.New("更新失败")
	}

	// 添加节点
	setPermission(req.Type, req.CheckedList, req.Title, req.Path, req.Id, userId)

	return rows, nil
}

func (s *menuService) Delete(ids string) (int64, error) {
	if utils.AppDebug() {
		return 0, errors.New("演示环境，暂无权限操作")
	}
	// 记录ID
	idsArr := strings.Split(ids, ",")
	if len(idsArr) == 1 {
		// 单个删除
		entity := model.Menu{Id: gconv.Int(ids)}
		rows, err := entity.Delete()
		if err != nil || rows == 0 {
			return 0, err
		}
		return rows, nil
	} else {
		// 批量删除
		count := 0
		for _, v := range idsArr {
			id, _ := strconv.Atoi(v)
			entity := &model.Menu{Id: id}
			rows, err := entity.Delete()
			if rows == 0 || err != nil {
				continue
			}
			count++
		}
		return int64(count), nil
	}
}

// 添加节点
func setPermission(menuType int, checkedList []int, name string, url string, parentId int, userId int) {
	if menuType != 0 || len(checkedList) == 0 || url == "" {
		return
	}
	// 删除现有节点
	utils.XormDb.Where("parent_id=?", parentId).Delete(&model.Menu{})
	// 模块名称
	moduleTitle := gstr.Replace(name, "管理", "")
	// 创建权限节点
	urlArr := strings.Split(url, "/")

	if len(urlArr) >= 3 {
		// 模块名
		moduleName := urlArr[len(urlArr)-1]
		// 节点处理
		for _, v := range checkedList {
			// 实例化对象
			var entity model.Menu
			// 节点索引
			value := gconv.Int(v)
			if value == 1 {
				entity.Title = "查询" + moduleTitle
				entity.Path = "/" + moduleName + "/list"
				entity.Permission = "sys:" + moduleName + ":list"
			} else if value == 5 {
				entity.Title = "添加" + moduleTitle
				entity.Path = "/" + moduleName + "/add"
				entity.Permission = "sys:" + moduleName + ":add"
			} else if value == 10 {
				entity.Title = "修改" + moduleTitle
				entity.Path = "/" + moduleName + "/update"
				entity.Permission = "sys:" + moduleName + ":update"
			} else if value == 15 {
				entity.Title = "删除" + moduleTitle
				entity.Path = "/" + moduleName + "/delete"
				entity.Permission = "sys:" + moduleName + ":delete"
			} else if value == 20 {
				entity.Title = moduleTitle + "详情"
				entity.Path = "/" + moduleName + "/detail"
				entity.Permission = "sys:" + moduleName + ":detail"
			} else if value == 25 {
				entity.Title = "设置状态"
				entity.Path = "/" + moduleName + "/status"
				entity.Permission = "sys:" + moduleName + ":status"
			} else if value == 30 {
				entity.Title = "批量删除"
				entity.Path = "/" + moduleName + "/dall"
				entity.Permission = "sys:" + moduleName + ":dall"
			} else if value == 35 {
				entity.Title = "添加子级"
				entity.Path = "/" + moduleName + "/addz"
				entity.Permission = "sys:" + moduleName + ":addz"
			} else if value == 40 {
				entity.Title = "全部展开"
				entity.Path = "/" + moduleName + "/expand"
				entity.Permission = "sys:" + moduleName + ":expand"
			} else if value == 45 {
				entity.Title = "全部折叠"
				entity.Path = "/" + moduleName + "/collapse"
				entity.Permission = "sys:" + moduleName + ":collapse"
			} else if value == 50 {
				entity.Title = "导出" + moduleTitle
				entity.Path = "/" + moduleName + "/export"
				entity.Permission = "sys:" + moduleName + ":export"
			} else if value == 55 {
				entity.Title = "导入" + moduleTitle
				entity.Path = "/" + moduleName + "/import"
				entity.Permission = "sys:" + moduleName + ":import"
			} else if value == 60 {
				entity.Title = "分配权限"
				entity.Path = "/" + moduleName + "/permission"
				entity.Permission = "sys:" + moduleName + ":permission"
			} else if value == 65 {
				entity.Title = "重置密码"
				entity.Path = "/" + moduleName + "/resetPwd"
				entity.Permission = "sys:" + moduleName + ":resetPwd"
			}
			entity.ParentId = parentId
			entity.Type = 1
			entity.Status = 1
			entity.Target = "_self"
			entity.Sort = value
			entity.CreateUser = userId
			entity.CreateTime = time.Now().Unix()
			entity.UpdateUser = userId
			entity.UpdateTime = time.Now().Unix()
			entity.Mark = 1

			// 插入节点
			entity.Insert()
		}
	}
}

// 获取菜单权限列表
func (s *menuService) GetPermissionList(userId int) interface{} {
	if userId == 1 {
		// 管理员(拥有全部权限)
		menuList, _ := Menu.GetTreeList()
		return menuList
	} else {
		// 非管理员

		// 数据转换
		list := make([]model.Menu, 0)
		// 查询数据
		utils.XormDb.Table("sys_menu").Alias("m").
			Join("INNER", []string{"sys_role_menu", "r"}, "m.id = r.menu_id").
			Join("INNER", []string{"sys_user_role", "ur"}, "ur.role_id=r.role_id").
			Where("ur.user_id=? AND m.type=0 AND m.`status`=1 AND m.mark=1", userId).
			Cols("m.*").
			OrderBy("m.id asc").
			Find(&list)

		// 数据处理
		var menuNode vo.MenuTreeNode
		makeTree(list, &menuNode)
		return menuNode.Children
	}
}

// 获取子级菜单
func (s *menuService) GetTreeList() ([]*vo.MenuTreeNode, error) {
	var menuNode vo.MenuTreeNode
	list := make([]model.Menu, 0)
	err := utils.XormDb.Where("type=0 and mark=1").OrderBy("sort").Find(&list)
	if err != nil {
		return nil, err
	}
	makeTree(list, &menuNode)
	return menuNode.Children, nil
}

//递归生成分类列表
func makeTree(menu []model.Menu, tn *vo.MenuTreeNode) {
	for _, c := range menu {
		if c.ParentId == tn.Id {
			child := &vo.MenuTreeNode{}
			child.Menu = c
			tn.Children = append(tn.Children, child)
			makeTree(menu, child)
		}
	}
}

// 获取权限节点列表
func (s *menuService) GetPermissionsList(userId int) []string {
	if userId == 1 {
		// 管理员,管理员拥有全部权限
		permissionList := make([]string, 0)
		utils.XormDb.Table("sys_menu").Cols("permission").Where("type=1").Where("mark=1").Find(&permissionList)
		return permissionList
	} else {
		// 非管理员
		permissionList := make([]string, 0)
		utils.XormDb.Table("sys_menu").Alias("m").
			Join("INNER", []string{"sys_role_menu", "r"}, "m.id = r.menu_id").
			Join("INNER", []string{"sys_user_role", "ur"}, "ur.role_id=r.role_id").
			Where("ur.user_id=? AND m.type=1 AND m.`status`=1 AND m.mark=1", userId).
			Cols("m.permission").
			Find(&permissionList)
		return permissionList
	}
}
