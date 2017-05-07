package reddit

import (
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type Subreddits struct {
	Default []string `yaml:"default"`
	Custom  []struct {
		Chats      []int64  `yaml:"chats"`
		Subreddits []string `yaml:"subreddits"`
	} `yaml:"custom"`
	Chats map[int64][]string
}

func NewSubreddits(filename *string) (*Subreddits, error) {
	filepath, _ := filepath.Abs(*filename)
	file, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Could not read file %s: %s", filepath, err.Error()))
	}

	var subreddits Subreddits
	err = yaml.Unmarshal(file, &subreddits)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Could not parse yaml: %s", err.Error()))
	}

	subreddits.Chats = make(map[int64][]string)

	for _, custom := range subreddits.Custom {
		for _, chat := range custom.Chats {
			subreddits.Chats[chat] = custom.Subreddits
		}
	}

	return &subreddits, nil
}

func (s *Subreddits) PickRandom(chatId int64) string {
	var list []string
	if chatList, ok := s.Chats[chatId]; ok {
		list = chatList
	} else {
		list = s.Default
	}

	return list[rand.Intn(len(list))]
}
