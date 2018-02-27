package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"

	pb "github.com/Jeiwan/micros/post/proto/post"
	"github.com/hashicorp/consul/api"
	uuid "github.com/satori/go.uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type storage interface {
	Create(*pb.Post) (*pb.Post, error)
	List() ([]*pb.Post, error)
	Get(int64) (*pb.Post, error)
}

type postStorage struct {
	posts []*pb.Post
}

func (s *postStorage) Create(post *pb.Post) (*pb.Post, error) {
	post.Id = int64(len(s.posts) + 1)
	s.posts = append(s.posts, post)
	return post, nil
}

func (s *postStorage) List() ([]*pb.Post, error) {
	return s.posts, nil
}

func (s *postStorage) Get(postID int64) (*pb.Post, error) {
	for _, post := range s.posts {
		if post.Id == postID {
			return post, nil
		}
	}

	return nil, errors.New("Not found")
}

type service struct {
	storage storage
}

func (s *service) CreatePost(ctx context.Context, req *pb.Post) (*pb.Response, error) {
	post, err := s.storage.Create(req)
	if err != nil {
		log.Fatal(err)
	}

	return &pb.Response{Status: true, Post: post}, nil
}

func (s *service) ListPosts(ctx context.Context, req *pb.ListRequest) (*pb.Response, error) {
	posts, err := s.storage.List()
	if err != nil {
		log.Fatal(err)
	}

	return &pb.Response{
		Status: true,
		Posts:  posts,
	}, nil
}

func (s *service) GetPost(ctx context.Context, req *pb.GetRequest) (*pb.Response, error) {
	post, err := s.storage.Get(req.PostID)
	if err != nil {
		return nil, err
	}

	resp := &pb.Response{
		Status: true,
		Post:   post,
	}
	return resp, nil
}

func main() {
	storage := &postStorage{}

	host := os.Getenv("HOSTNAME")
	port := os.Getenv("PORT")
	portInt, err := strconv.Atoi(port)
	if err != nil {
		log.Fatal(err)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%s", host, port))
	if err != nil {
		log.Fatal(err)
	}

	consul, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		log.Fatal(err)
	}

	serviceReg := &api.AgentServiceRegistration{
		ID:      fmt.Sprintf("post-%s", uuid.NewV4().String()),
		Name:    "post",
		Tags:    []string{},
		Address: host,
		Port:    portInt,
	}

	err = consul.Agent().ServiceRegister(serviceReg)
	if err != nil {
		log.Fatal(err)
	}
	defer consul.Agent().ServiceDeregister(serviceReg.ID)

	s := grpc.NewServer()

	pb.RegisterPostServiceServer(s, &service{storage})

	reflection.Register(s)

	fmt.Println("Starting the gRPC server on", fmt.Sprintf("%s:%s", host, port))
	if err := s.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
