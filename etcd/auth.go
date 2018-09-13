package etcd

import (
	"context"
	"errors"

	"github.com/coreos/etcd/auth/authpb"
	"github.com/coreos/etcd/clientv3"
)

var (
	EtcdNoAuthClientError = errors.New("etcd auth client is nil")
)

type Permission struct {
	Key      string
	RangeEnd string
	Type     authpb.Permission_Type
}

/**
 *
 * perms := []*Permission{
		&Permission{Key:"/warden/", RangeEnd:"/warden/z", Type: clientv3.PermReadWrite},
		&Permission{Key:"/web/", RangeEnd:"/web/z", Type: clientv3.PermReadWrite},
	}
	if err := createRoleWithPermission("web", perms , authAPI); err != nil {
		log.Fatal(err)
	}
 *
*/
func CreateRoleWithPermission(role string, perms []*Permission) (err error) {
	if authClient == nil {
		return EtcdNoAuthClientError
	}
	resp, err := authClient.RoleAdd(context.TODO(), role)
	if err != nil {
		fmt.Error("etcd:auth add role error[%v][%q]\n", err, resp)
		return err
	}

	for _, perm := range perms {
		if _, err = authClient.RoleGrantPermission(
			context.TODO(),
			role,          // role name
			perm.Key,      // key
			perm.RangeEnd, // range end
			clientv3.PermissionType(perm.Type),
		); err != nil {
			return err
		}
	}
	return
}

/**
 *
 * 新增用户
 */
func AddUser(role, user, pass string) error {
	if authClient == nil {
		return EtcdNoAuthClientError
	}
	//添加一个用户
	if _, err := authClient.UserAdd(context.TODO(), user, pass); err != nil {
		return err
	}
	return nil
}

/**
 *
 * 给用户赋予权限
 */
func GrantUser(user, role string) error {
	if _, err := authClient.UserGrantRole(context.TODO(), user, role); err != nil {
		return err
	}
	return nil
}
