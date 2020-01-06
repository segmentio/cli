package cli_test

import (
	"os"

	"github.com/segmentio/cli"
)

func ExampleFormat_json() {
	cmd := cli.Command(func() error {
		p, err := cli.Format("json", os.Stdout)
		if err != nil {
			return err
		}
		defer p.Flush()

		p.Print(struct {
			Message string
		}{"Hello World!"})
		return nil
	})

	cli.Call(cmd)
	// Output:
	// {
	//   "Message": "Hello World!"
	// }
}

func ExampleFormat_yaml() {
	cmd := cli.Command(func() error {
		p, err := cli.Format("yaml", os.Stdout)
		if err != nil {
			return err
		}
		defer p.Flush()

		type output struct {
			Value int `json:"value"`
		}

		p.Print(output{1})
		p.Print(output{2})
		p.Print(output{3})
		return nil
	})

	cli.Call(cmd)
	// Output:
	// value: 1
	// ---
	// value: 2
	// ---
	// value: 3
}

func ExampleFormat_text_string() {
	cmd := cli.Command(func() error {
		p, err := cli.Format("text", os.Stdout)
		if err != nil {
			return err
		}
		defer p.Flush()

		p.Print("hello")
		p.Print("world")
		return nil
	})

	cli.Call(cmd)
	// Output:
	// hello
	// world
}

func ExampleFormat_text_struct() {
	cmd := cli.Command(func() error {
		p, err := cli.Format("text", os.Stdout)
		if err != nil {
			return err
		}
		defer p.Flush()

		type output struct {
			ID    string
			Name  string
			Value int
		}

		p.Print(output{"1234", "A", 1})
		p.Print(output{"5678", "B", 2})
		p.Print(output{"9012", "C", 3})
		return nil
	})

	cli.Call(cmd)
	// Output:
	// ID    NAME  VALUE
	// 1234  A     1
	// 5678  B     2
	// 9012  C     3
}

func ExampleFormat_text_map() {
	cmd := cli.Command(func() error {
		p, err := cli.Format("text", os.Stdout)
		if err != nil {
			return err
		}
		defer p.Flush()

		p.Print(map[string]interface{}{
			"ID":    "1234",
			"Name":  "A",
			"Value": 1,
		})

		p.Print(map[string]interface{}{
			"ID":    "5678",
			"Name":  "B",
			"Value": 2,
		})

		p.Print(map[string]interface{}{
			"ID":    "9012",
			"Name":  "C",
			"Value": 3,
		})
		return nil
	})

	cli.Call(cmd)
	// Output:
	// ID    NAME  VALUE
	// 1234  A     1
	// 5678  B     2
	// 9012  C     3
}

func ExampleFormatList_json() {
	cmd := cli.Command(func() error {
		p, err := cli.FormatList("json", os.Stdout)
		if err != nil {
			return err
		}
		defer p.Flush()

		p.Print(struct {
			Message string
		}{"Hello World!"})
		return nil
	})

	cli.Call(cmd)
	// Output:
	// [
	//   {
	//     "Message": "Hello World!"
	//   }
	// ]
}

func ExampleFormatList_yaml() {
	cmd := cli.Command(func() error {
		p, err := cli.FormatList("yaml", os.Stdout)
		if err != nil {
			return err
		}
		defer p.Flush()

		type output struct {
			Value int `json:"value"`
		}

		p.Print(output{1})
		p.Print(output{2})
		p.Print(output{3})
		return nil
	})

	cli.Call(cmd)
	// Output:
	// - value: 1
	// - value: 2
	// - value: 3
}
