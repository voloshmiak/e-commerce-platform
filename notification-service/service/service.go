package service

import "fmt"

type NotificationService struct{}

func (s *NotificationService) SendNotification(message string) {
	fmt.Println(message)
}
