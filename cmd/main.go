package main

import (
	"fmt"
	"guimox/internal/whatsapp"
)

func main() {
	info, err := whatsapp.GetNewsletterData()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println(info)
	return
}
