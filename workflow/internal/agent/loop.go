package agent

import "fmt"

func (a *Agent) Run(prompt string) error {
	a.Messages = append(a.Messages, Message{
		Role:    "user",
		Content: prompt,
	})
	fmt.Println("received:", prompt)
	return nil
}
