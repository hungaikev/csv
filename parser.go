package  main

import (
	"bufio"
	"container/list"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/tealeg/xlsx"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func ReadCsvFile(filepath string) []map[string]interface{} {
	// Load a csv file.
	f, _ := os.Open(filepath)
	// Create a new reader.
	r := csv.NewReader(bufio.NewReader(f))
	result, _ := r.ReadAll()
	parsedData := make([]map[string]interface{}, 0, 0)
	header_name := result[0]

	for row_counter, row := range result {
		if row_counter != 0 {
			var singleMap = make(map[string]interface{})
			for col_counter, col := range row {
				singleMap[header_name[col_counter]] = col
			}
			if len(singleMap) > 0 {
				parsedData = append(parsedData, singleMap)
			}
		}
	}
	fmt.Println("Length of parsedData: ", len(parsedData))
	return parsedData
}

func ReadXlsFile(filepath string) []map[string]interface{} {
	xlFile, err := xlsx.OpenFile(filepath)
	if err != nil {
		fmt.Println("Error reading the file")
	}

	parsedData := make([]map[string]interface{},0,0)
	header_name := list.New()
	//Sheet
	for _, sheet := range xlFile.Sheets {
		// rows
		for row_counter, row := range sheet.Rows {
			// column
			header_iterator := header_name.Front()
			var singleMap = make(map[string]interface{})
			for _, cell := range row.Cells {
				if row_counter == 0 {
					text := cell.String()
					header_name.PushBack(text)
				} else {
					text := cell.String()
					singleMap[header_iterator.Value.(string)] = text
					header_iterator = header_iterator.Next()
				}
				if row_counter != 0 && len(singleMap) > 0 {
					parsedData = append(parsedData,singleMap)
				}
			}
		}
	}
	fmt.Println("Length of parsedData: ", len(parsedData))
	return parsedData
}

func ExcelCsvParser(blobPath string, blobExtension string)(parsedData []map[string]interface{}) {
	fmt.Println(" -----------------------> We are in product.go")
	if blobExtension == ".csv" {
		fmt.Println(" -----------------------------We are parsing a csv file. -----------------")
		parsedData := ReadCsvFile(blobPath)
		fmt.Printf("Type:%T\n", parsedData)
		return parsedData
	} else if blobExtension == ".xlsx" {
		fmt.Println("------------------------We are parsing an xlsx file. ------------------")
		parsedData := ReadXlsFile(blobPath)
		return parsedData
	}
	return  parsedData
}

func uploadData(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		fmt.Println("GET")
		t, _ := template.ParseFiles("./templates/index.html")
		t.Execute(w, nil)
	} else if r.Method == "POST" {
		fmt.Println("POST")
		file, handler, err := r.FormFile("uploadfile")
		defer file.Close()
		if err != nil {
			log.Printf("Error while Posting data")
			t, _ := template.ParseFiles("./templates/index.html")
			t.Execute(w, nil)
		} else {
			fmt.Println("error throws in else statement")
			fmt.Println("handler.Filename",handler.Filename)
			fmt.Printf("Type of handler.Filename:%T\n",handler.Filename)
			fmt.Println("Length:",len(handler.Filename))
			f, err := os.OpenFile("./data/"+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
			if err != nil {
				fmt.Println("Error:",err)
				t, _ := template.ParseFiles("./templates/index.html")
				t.Execute(w, nil)
			}
			defer f.Close()
			io.Copy(f, file)
			blobPath := "./data/" + handler.Filename
			var extension = filepath.Ext(blobPath)
			parsedData := ExcelCsvParser(blobPath, extension)
			parsedJson, _ := json.Marshal(parsedData)
			fmt.Println(string(parsedJson))
			err = os.Remove(blobPath)
			if(err!=nil){
				fmt.Println(err.Error())
			}else{
				fmt.Println("File has been deleted successfully.")
			}
			t, _ := template.ParseFiles("./templates/index.html")
			t.Execute(w, string(parsedJson))

		}

	} else {
		log.Printf("Error while Posting Data")
		t, _ := template.ParseFiles("./templates/index.html")
		t.Execute(w, nil )
	}
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/", uploadData)
	// http.FileServer(http.Dir("./templates"))
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./templates/")))
	log.Fatal(http.ListenAndServe(":9000", router))

}