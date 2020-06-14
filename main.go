package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	_ "github.com/lib/pq"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)
//https://www.alphavantage.co/query?function=CURRENCY_EXCHANGE_RATE&from_currency=BTC&to_currency=CNY&apikey=
const (
	CONN_HOST = "localhost"
	CONN_PORT = "8080"
	URL_1 = "https://alpha-vantage.p.rapidapi.com/query?from_currency="
	URL_2 = "&function=CURRENCY_EXCHANGE_RATE&to_currency="
	API_KEY="TGNFWBLDDA6D8B7U"
	URL_JSON_1="https://www.alphavantage.co/query?function=CURRENCY_EXCHANGE_RATE&from_currency="
	URL_JSON_2="&to_currency="
	URL_JSON_3="&apikey="
	STOCK_URL_1="https://www.alphavantage.co/query?function=GLOBAL_QUOTE&symbol="
	STOCK_URL_2="&apikey="
	HOST="ec2-18-209-187-54.compute-1.amazonaws.com"
	PORT=5432
	USER="tmtwjudawuxoxu"
	PASSWORD="741dee31b5e5aa44e84c02bc2bcfb072730a24b92d032d2703ab9bb69586b7ea"
	DB_NAME="d5cm8frgt31r8u"
)

type User struct {
	UserName string
	Password string
	Email string
}
type Stock struct {
	Symbol string
	Quantity int
	OrigValue float64
}
type Stocks []Stock
type Crypto struct {
	Currency string
	Quantity float64
	Orig_value float64
}
type Cryptos []Crypto

type Convert struct {
	Cryptocurrency string
	Currency string
}

type AutoGenerated struct {
	RealtimeCurrencyExchangeRate struct {
		FromCurrencyCode string `json:"1. From_Currency Code"`
		FromCurrencyName string `json:"2. From_Currency Name"`
		ToCurrencyCode string `json:"3. To_Currency Code"`
		ToCurrencyName  string `json:"4. To_Currency Name"`
		ExchangeRate    string `json:"5. Exchange Rate"`
		LastRefreshed    string `json:"6. Last Refreshed"`
		TimeZone       string `json:"7. Time Zone"`
		BidPrice       string `json:"8. Bid Price"`
		AskPrice        string `json:"9. Ask Price"`
	} `json:"Realtime Currency Exchange Rate"`
}
type StockStruct struct {
	GlobalQuote struct {
		Symbol string `json:"01. symbol"`
		Open string `json:"02. open"`
		High string `json:"03. high"`
		Low  string `json:"04. low"`
		Price string `json:"05. price"`
		Volume string `json:"06. volume"`
		LatestTradingDay string `json:"07. latest trading day"`
		PreviousClose string `json:"08. previous close"`
		Change string `json:"09. change"`
		ChangePercent string `json:"10. change percent"`
	} `json:"Global Quote"`
}


type AllData struct {
	Id int
	ExchangeRate string
	ToCurrencyName  string
	FromCurrencyName string
}
var psqlInfo string
var db *sql.DB
func init()  {
	psqlInfo = fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=require",
		HOST, PORT, USER, PASSWORD, DB_NAME)
}

func main()  {
	var err error
	db,err  =sql.Open("postgres",psqlInfo)
	if err!=nil {
		panic(err)
	}
	defer db.Close()
	err = db.Ping()
	if err != nil {
		panic(err)
	}
	log.Printf("%T",db)
	router :=mux.NewRouter()
	router.HandleFunc("/parser",infoAccesor).Methods("POST")
	router.HandleFunc("/insertTest",testDB).Methods("GET")
	router.HandleFunc("/",infoForm).Methods("GET")
	//router.HandleFunc("/cryptoExchange",infoProvider).Methods
	router.HandleFunc("/register",userRegisterForm).Methods("GET")
	router.HandleFunc("/addData",currencyRegisterForm).Methods("GET")
	router.HandleFunc("/addData",currencyRegister).Methods("POST")
	router.HandleFunc("/register",userRegisteration).Methods("POST")
	router.HandleFunc("/cryptoExchangeJSON",jsonProvider).Methods("GET")
	router.HandleFunc("/profit",profitCalculator).Methods("GET")
	router.HandleFunc("/stockPrice",stockCalculator).Methods("GET")
	router.HandleFunc("/stockForm",stockRegistration).Methods("GET")
	router.HandleFunc("/stockForm",stockProcess).Methods("POST")
	router.HandleFunc("/stockProfit",stockProfit).Methods("GET")


	http.ListenAndServe(CONN_HOST+":"+CONN_PORT,router)
}

func stockProfit(writer http.ResponseWriter, request *http.Request) {
	sqlStatement := `SELECT stock,quantity,orig_value FROM stocks WHERE users_id=1`
	rows, err := db.Query(sqlStatement)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	var stocks Stocks
	for rows.Next() {
		var symbol string
		var quantity int
		var orig_value float64
		err = rows.Scan(&symbol, &quantity, &orig_value)
		if err != nil {
			panic(err)
		}
		var stock = Stock{
			Symbol:    symbol,
			Quantity:  quantity,
			OrigValue: orig_value,
		}
		stocks = append(stocks, stock)
	}
	err = rows.Err()
	if err != nil {
		panic(err)
	}
	var profit float64 = 0
	for _,stock:=range stocks{
		itemProfit:=(float64(stock.Quantity))*(StockRate(stock.Symbol)-stock.OrigValue)
		fmt.Fprintf(writer,"profit for "+stock.Symbol+" equals "+fmt.Sprintf("%f", itemProfit)+"\n")
		profit+=profit+itemProfit
	}
	fmt.Fprintf(writer,"net profit is "+fmt.Sprintf("%f", profit))


}







func stockProcess(writer http.ResponseWriter, request *http.Request) {
	request.ParseForm()
	stock :=new(Stock)
	decoder:=schema.NewDecoder()
	decodeErr:=decoder.Decode(stock,request.PostForm)
	if decodeErr != nil {
		log.Printf("error in decoding form",decodeErr)
	}
	sqlStatement:=`
    INSERT INTO stocks (stock,quantity,orig_value,users_id)
	VALUES ($1,$2,$3,$4)`
	_,err:=db.Exec(sqlStatement,stock.Symbol,stock.Quantity,stock.OrigValue,1)
	if err != nil {
		log.Printf("error in database",err)
	}else {
		fmt.Fprintf(writer," stock transaction  successfully registered")
	}
}

func stockRegistration(writer http.ResponseWriter, request *http.Request) {
	parsedTemplate,err:=template.ParseFiles("templates/stockForm.html")
	if err!=nil {
		log.Printf("error parsing stock form",err)
		return
	}
	err = parsedTemplate.Execute(writer,nil)
	if err != nil {
		log.Printf("error executing",err)
		return
	}
}

func stockCalculator(writer http.ResponseWriter, request *http.Request) {
	vars:=request.URL.Query()
	symbol,symbolOk:=vars["symbol"]
	if !symbolOk{
		log.Printf("error getting symbol")
		return
	}
	price :=StockRate(symbol[0])
	fmt.Fprintf(writer,"price of "+symbol[0]+" is "+fmt.Sprintf("%f", price))
}

func profitCalculator(writer http.ResponseWriter, request *http.Request) {

	sqlStatement:=`SELECT curr,quantity,orig_value FROM transaction_store WHERE users_id=1`
	rows,err:=db.Query(sqlStatement)
	if err!=nil {
		panic(err)
	}
	defer rows.Close()
	var cryptos Cryptos
	for rows.Next(){
		var curr string
		var quantity float64
		var orig_value float64
		err = rows.Scan(&curr,&quantity,&orig_value)
		if err!=nil {
			panic(err)
		}
		var crypto =Crypto{
			Currency:   curr,
			Quantity:   quantity,
			Orig_value: orig_value,
		}
		cryptos = append(cryptos,crypto)

	}
	err = rows.Err()
	if err != nil {
		panic(err)
	}
	var profit float64 = 0
	for _,crypt:=range cryptos{
		itemProfit:=(crypt.Quantity)*(getExchangeRate(crypt.Currency,"INR")-crypt.Orig_value)
		fmt.Fprintf(writer,"profit for "+crypt.Currency+"equals "+fmt.Sprintf("%f", itemProfit)+"\n")
		profit+=profit+itemProfit
	}
	fmt.Fprintf(writer,"net profit is "+fmt.Sprintf("%f", profit))


}

func currencyRegister(writer http.ResponseWriter, request *http.Request) {
	request.ParseForm()
	crypto:=new(Crypto)
	decoder :=schema.NewDecoder()
	decodeErr:=decoder.Decode(crypto,request.PostForm)
	if decodeErr != nil {
		log.Printf("error in decoding form",decodeErr)
	}
	sqlStatement:=`
    INSERT INTO transaction_store (curr,quantity,orig_value,users_id)
	VALUES ($1,$2,$3,$4)`
	_,err:=db.Exec(sqlStatement,crypto.Currency,crypto.Quantity,crypto.Orig_value,1)
	if err != nil {
		log.Printf("error in database",err)
	}else {
		fmt.Fprintf(writer,"transaction  successfully registered")
	}
}



func currencyRegisterForm(writer http.ResponseWriter, request *http.Request) {
	parsedTemplate,err:=template.ParseFiles("templates/addCurr.html")
	if err != nil {
		log.Printf("error parsing currency form",err)
		return
	}
	err =parsedTemplate.Execute(writer,nil)
	if err!=nil {
		log.Printf("error executing",err)
	}
}

func userRegisteration(writer http.ResponseWriter, request *http.Request) {
	request.ParseForm()
	user :=new(User)
	decoder:=schema.NewDecoder()
	decodeErr:=decoder.Decode(user,request.PostForm)
	if decodeErr != nil {
		log.Printf("error in decoding form",decodeErr)
	}
	sqlStatement:=`
    INSERT INTO users (username,password,email)
	VALUES ($1,$2,$3)`
	_,err:=db.Exec(sqlStatement,user.UserName,user.Password,user.Email)
	if err != nil {
		log.Printf("error in database",err)
	}else {
		fmt.Fprintf(writer,"user successfully registered"+user.UserName)
	}
}

func userRegisterForm(writer http.ResponseWriter, request *http.Request) {
	parsedTemplate,err :=template.ParseFiles("templates/register.html")
	if err!=nil {
		log.Printf("error in parsing html",err)
		return
	}
	err =parsedTemplate.Execute(writer,nil)
	if err != nil {
		log.Printf("error in executing",err)
		return
	}
}

func testDB(writer http.ResponseWriter, request *http.Request) {
	sqlStatement := `
	INSERT INTO users (username,password,email)
	VALUES ('kkiu', 'jij','Calhoun@gmail.com')`
	/**INSERT INTO RateStore(FromCurr,ToCurr,Rate)
	VALUES ('BTC','INR',89)**/
	_,err :=db.Exec(sqlStatement)
	if err != nil {
		panic(err)
	} else{
		fmt.Fprintf(writer,"successfully inserted in DB")
	}


}

func jsonProvider(writer http.ResponseWriter, request *http.Request) {
	vars :=request.URL.Query()
	crypto, crOk :=vars["crypto"]
	currency,cuOK:=vars["cur"]
	apiurl:=""
	if cuOK&&crOk{
		apiurl=URL_JSON_1+crypto[0]+URL_JSON_2+currency[0]+URL_JSON_3+API_KEY
	}else{
		log.Printf("wrong currency entered")
		return
	}
	response,err:=http.Get(apiurl)
	defer response.Body.Close()
	if err!=nil {
		log.Fatal("error getting response",err)
	}

	responseData,err:=ioutil.ReadAll(response.Body)

	if err != nil {
		log.Fatal(err)
	}
	var resp AutoGenerated
	json.Unmarshal([]byte(responseData),&resp)


	parsedFile, _ := template.ParseFiles("templates/infoShower.html")
	allData := AllData{
		ExchangeRate:     resp.RealtimeCurrencyExchangeRate.ExchangeRate,
		ToCurrencyName:   resp.RealtimeCurrencyExchangeRate.ToCurrencyName,
		FromCurrencyName: resp.RealtimeCurrencyExchangeRate.FromCurrencyName,
	}
	parsedFile.Execute(writer,allData)




}

func infoForm(writer http.ResponseWriter, request *http.Request) {
	parsedFile,_ :=template.ParseFiles("templates/cryptoForm.html")
	err :=parsedFile.Execute(writer,nil)
	if err!=nil {
		log.Fatal("error parsing html form",err)
	}

}

func infoAccesor(writer http.ResponseWriter, request *http.Request) {
	log.Printf("inside info accessor")
	request.ParseForm()
	convert :=new (Convert)
	decoder :=schema.NewDecoder()
	decodeErr:=decoder.Decode(convert,request.PostForm)
	if decodeErr !=nil{
		log.Fatal("error reading form",decodeErr)
	}
	log.Printf("reading form for "+convert.Cryptocurrency+" and "+convert.Currency)
	urlString := "http://localhost:8080/cryptoExchangeJSON?crypto="+convert.Cryptocurrency+"&cur="+convert.Currency
	//urlString ="https://www.google.co.in/"
	http.Redirect(writer,request,urlString,http.StatusMovedPermanently)
}

/**func infoProvider(writer http.ResponseWriter, request *http.Request) {
	log.Printf("inside info provider")
	vars :=request.URL.Query()
	crypto, crOk :=vars["crypto"]
	currency,cuOK:=vars["cur"]
	apiurl := ""

	if crOk&&cuOK {
		log.Printf("making url for "+crypto[0]+"and"+currency[0])
		apiurl = URL_1 + crypto[0] + URL_2 + currency[0]
	} else{
		log.Fatal("error reading arguments")
	}
	log.Printf(apiurl)
	info :=apiCall(apiurl)
	fmt.Fprintf(writer,info)
}
**/
func apiCall(apiUrl string)string{
	info:=""
	req,_ :=http.NewRequest("GET",apiUrl,nil)
	req.Header.Add("x-rapidapi-host", "alpha-vantage.p.rapidapi.com")
	req.Header.Add("x-rapidapi-key","a9935374bemshb0147f63949e0e1p17c9d1jsn6a305965d763")

	res,_ :=http.DefaultClient.Do(req)
	defer res.Body.Close()
	body ,_:=ioutil.ReadAll(res.Body)
	info = string(body)
	log.Printf("%T",res.Body)

	return info
}

func getExchangeRate(toCurr string,fromCurr string)float64{

	apiurl:=URL_JSON_1+fromCurr+URL_JSON_2+toCurr+URL_JSON_3+API_KEY
	response,err:=http.Get(apiurl)
	defer response.Body.Close()
	if err!=nil {
		log.Fatal("error getting response",err)
	}

	responseData,err:=ioutil.ReadAll(response.Body)

	if err != nil {
		log.Fatal(err)
	}
	var resp AutoGenerated
	json.Unmarshal([]byte(responseData),&resp)
	rate := resp.RealtimeCurrencyExchangeRate.ExchangeRate
	s, err := strconv.ParseFloat(rate, 32)

	return s
}

func StockRate(stockSymbol string)float64{

	apiurl:=STOCK_URL_1+stockSymbol+STOCK_URL_2+API_KEY
	response,err:=http.Get(apiurl)
	defer response.Body.Close()
	if err!=nil {
		log.Fatal("error getting response",err)
	}

	responseData,err:=ioutil.ReadAll(response.Body)

	if err != nil {
		log.Fatal(err)
	}
	var resp StockStruct
	json.Unmarshal([]byte(responseData),&resp)
	rateString := resp.GlobalQuote.Price
	rate, _ :=strconv.ParseFloat(rateString,32)


	//log.Printf(fmt.Sprintf("%F",rate))

	return rate
}