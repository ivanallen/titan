package archiver

import "log"

func ExampleZip() {
	err := Zip("./tmp", "./tmp.zip")
	if err != nil {
		log.Printf("zip error:%v", err)
	}
	// Output:
}
