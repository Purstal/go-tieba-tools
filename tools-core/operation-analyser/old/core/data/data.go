package data

import "time"

type Data struct {
	ThreadCount int
	PostCount   int
	SameThread  map[int][]*Post
	SameAccount map[string][]*Post
	OldPosts    map[*Post]*time.Time

	ChThreadCount chan int
	ChPostCount   chan int
	ChSameThread  chan struct {
		Tid int
		Ptr *Post
	}
	ChSameAccount chan struct {
		Un  string
		Ptr *Post
	}
	ChOldPosts chan struct {
		Ptr      *Post
		Interval *time.Time
	}
}

type Account struct {
	username string
}
type Post struct {
	PostType int
	Title    string
	Content  string
	Author   string
}

func NewData() *Data {
	var data Data

	go func() {
		for {
			t := <-data.ChSameThread
			data.SameThread[t.Tid] = append(data.SameThread[t.Tid], t.Ptr)
		}

	}()

	go func() {
		for {
			t := <-data.ChSameAccount
			data.SameAccount[t.Un] = append(data.SameAccount[t.Un], t.Ptr)
		}

	}()
	go func() {
		for {
			t := <-data.ChOldPosts
			if t.Interval == nil || t.Interval.Before(*data.OldPosts[t.Ptr]) {
				data.OldPosts[t.Ptr] = t.Interval
			}

		}

	}()
	go func() {
		for {
			data.ThreadCount += <-data.ChThreadCount

		}
	}()
	go func() {
		for {
			data.PostCount += <-data.ChPostCount
		}
	}()

	return &data
}
