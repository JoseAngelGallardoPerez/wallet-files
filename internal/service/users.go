package service

import (
	"context"
	"net/http"

	"github.com/Confialink/wallet-files/internal/srvdiscovery"

	pb "github.com/Confialink/wallet-users/rpc/proto/users"
)

type Users struct {
}

func NewUsers() *Users {
	return &Users{}
}

func (u *Users) GetByUID(uid string) (*pb.User, error) {
	req := pb.Request{UID: uid}
	client, err := u.getClient()
	if err != nil {
		return nil, err
	}
	resp, err := client.GetByUID(context.Background(), &req)
	if err != nil {
		return nil, err
	}
	return resp.User, nil
}

func (u *Users) UpdateProfileImageID(uid string, imageID uint64) error {
	req := pb.UpdateProfileImageIDRequest{
		UID:     uid,
		ImageID: imageID,
	}
	client, err := u.getClient()
	if err != nil {
		return err
	}
	_, err = client.UpdateProfileImageID(context.Background(), &req)
	return err
}

func (u *Users) getClient() (pb.UserHandler, error) {
	usersUrl, err := srvdiscovery.ResolveRPC(srvdiscovery.ServiceNameUsers)
	if nil != err {
		return nil, err
	}

	return pb.NewUserHandlerProtobufClient(usersUrl.String(), http.DefaultClient), nil
}
