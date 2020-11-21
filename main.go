package main
//бот дает ссылки на avito на написанные мной картины для продажи :)
//возможные запросы боту: пейзаж, купить картину маслом море, блин опять реклама, закат, саванна, купить что-нибудь   и т.д. по списку ключевых слов
//чтобы избежать повторной обработки id прошлого обработанного запроса перед запуском обновляю руками в main в строке 116 (как удобней сделать?)

import (
	"encoding/json"
	"fmt"
	"time"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

//описание картин
type PictureT struct {
	Name     string
	Path string
	KeyWords string
}

//структура, возвращаемая от getMe  набор result
type GetMeT struct {
	Ok     bool         `json:"ok"`
	Result GetMeResultT `json:"result"`
}

//структура, возвращаемая от getMe
type GetMeResultT struct {
	Id        int64  `json:"id"`
	IsBot     bool   `json:"is_bot"`
	FirstName string `json:"first_name"`
	Username  string `json:"username"`
}

//структура, возвращаемая от sendMessage
type SendMessageT struct {
	Ok     bool     `json:"ok"`
	Result MessageT `json:"result"`
}

//структура, возвращаемая от sendMessage набор result
type MessageT struct {
	MessageID int                          `json:"message_id"`
	From      GetUpdatesResultMessageFromT `json:"from"`
	Chat      GetUpdatesResultMessageChatT `json:"chat"`
	Date      int                          `json:"date"`
	Text      string                       `json:"text"`
}

//структура, возвращаемая от sendMessage внутри result набор from
type GetUpdatesResultMessageFromT struct {
	ID           int    `json:"id"`
	IsBot        bool   `json:"is_bot"`
	FirstName    string `json:"first_name"`
	Username     string `json:"username"`
	LanguageCode string `json:"language_code"`
}

//структура, возвращаемая от sendMessage внутри result набор chat
type GetUpdatesResultMessageChatT struct {
	ID        int    `json:"id"`
	FirstName string `json:"first_name"`
	Username  string `json:"username"`
	Type      string `json:"type"`
}

//структура, возвращаемая от getUpdates
type GetUpdatesT struct {
	Ok     bool                `json:"ok"`
	Result []GetUpdatesResultT `json:"result"`
}

//структура, возвращаемая от getUpdates набор result
type GetUpdatesResultT struct {
	UpdateID int                `json:"update_id"`
	Message  GetUpdatesMessageT `json:"message,omitempty"`
}

//структура, возвращаемая от getUpdates внутри result набор message
type GetUpdatesMessageT struct {
	MessageID int `json:"message_id"`
	From      struct {
		ID           int    `json:"id"`
		IsBot        bool   `json:"is_bot"`
		FirstName    string `json:"first_name"`
		LastName     string `json:"last_name"`
		Username     string `json:"username"`
		LanguageCode string `json:"language_code"`
	} `json:"from"`
	Chat struct {
		ID        int    `json:"id"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Username  string `json:"username"`
		Type      string `json:"type"`
	} `json:"chat"`
	Date int    `json:"date"`
	Text string `json:"text"`
}

const telegramBaseUrl = "https://api.telegram.org/bot"
const telegramToken = "1448230746:AAGv3g2bVwS5IIB_JAUQGjAY0dm2ihf8x7s"

const methodGetMe = "getMe"
const methodGetUpdates = "getUpdates"
const methodSendMessage = "sendMessage"
var lowerText string
var sendMessageTextNotEmpty bool 
var offsetUpdates int

func main() {

	//последний обработанный update_id пишем тут
	offsetUpdates = 561916160

	//картины, пути к ним и их ключевые слова
	pic0 := PictureT {"Возмущение", "А чего вы ждали? Оставили кодера без работы так надолго. Теперь картины пишу... https://vk.com/id75523881", "какого блин блиин"}
	pic1 := PictureT {"Волна", "https://www.avito.ru/sankt-peterburg/kollektsionirovanie/kartina_maslom_volna_50h70_2027475844", "волн море морск пейзаж"}
	pic2 := PictureT {"Две зебры", "https://www.avito.ru/sankt-peterburg/kollektsionirovanie/kartina_maslom_dve_zebry_30h40_2027558993", "зебр животн анималистик лошад кони пара"}
	pic3 := PictureT {"Закат в саванне.Львы", "https://www.avito.ru/sankt-peterburg/kollektsionirovanie/kartina_maslom_zakat_v_savanne._lvy_60h70_2027638233", "закат саван льв лев пейзаж небо страсть животн"}
	pic4 := PictureT {"Закат в саванне", "https://www.avito.ru/sankt-peterburg/mebel_i_interer/kartina_maslom_zakat_v_savanne_40h50sm_2027845921", "закат саван пейзаж небо страсть"}
	pic5 := PictureT {"Натюрморт с лошадкой", "https://www.avito.ru/sankt-peterburg/kollektsionirovanie/kartina_maslom_natyurmort_s_loshadkoy_18h24_2027607066", "натюрморт цветы детск игрушка уют цветами"}
	pic6 := PictureT {"Лимоны", "https://www.avito.ru/sankt-peterburg/mebel_i_interer/kartina_maslom_limony_2026997584", "натюрморт лимон кухн"}
	pic7 := PictureT {"Море ночью", "https://www.avito.ru/sankt-peterburg/mebel_i_interer/kartina_maslom_more_nochyu_20h25_sm_2027273204", "волн море морск пейзаж ночь ночн лун"}
	pic8 := PictureT {"Подсолнухи", "https://www.avito.ru/sankt-peterburg/kollektsionirovanie/kartina_maslom_podsolnuhi_60h70_2027066350", "цветы пейзаж подсолнух цветами"}
	pictures := [9] PictureT {pic0, pic1, pic2, pic3, pic4, pic5, pic6, pic7, pic8}

	// бесконечный цикл
	for {
		// получение сообщений
		body := getBodyByUrl(getUrlByMethod(methodGetUpdates))
		getUpdates := GetUpdatesT{}
		err := json.Unmarshal(body, &getUpdates)
		if err != nil {
			fmt.Printf("Error in unmarshal: %s", err.Error())

			return
		}

		// код обработки
		for _, update := range getUpdates.Result {
			
			//запоминаем максимум из уже обработанных update_id
			if update.UpdateID > offsetUpdates {
				offsetUpdates = update.UpdateID
			}
			fmt.Println("Обрабатываем запрос ",update.UpdateID, " offsetUpdates=", offsetUpdates)
			//готовим наши ответы
			lowerText = strings.ToLower(update.Message.Text)
			sendMessageTextNotEmpty = false
			//цикл по картинам
			for _, picture := range pictures {
				//цикл по всем ключам к данной картине
				for _, keyword := range strings.Split(picture.KeyWords, " ") {
					if strings.Contains(lowerText, strings.TrimSpace(keyword)) {
					
						url := getUrlByMethod(methodSendMessage)
						url = url + "?chat_id=" + strconv.Itoa(update.Message.Chat.ID) + "&text=" + picture.Path
			
						getBodyByUrl(url)
						
						sendMessageTextNotEmpty = true
						//fmt.Println("В запросе ",lowerText," для картины ",picture.Name," найдена подстрока ", keyword)
						break
					}
				}
			}
			//по умолчанию продаем "Волну"
			if !sendMessageTextNotEmpty && strings.Contains(lowerText, "купить") {
				url := getUrlByMethod(methodSendMessage)
				url = url + "?chat_id=" + strconv.Itoa(update.Message.Chat.ID) + "&text=" + "Да вот же хорошая картина! https://www.avito.ru/sankt-peterburg/kollektsionirovanie/kartina_maslom_volna_50h70_2027475844"
	
				getBodyByUrl(url)
				sendMessageTextNotEmpty = true
				continue
			}
			//ключевые слова в сообщении не встретились, посылаю рекламный пост 
			if !sendMessageTextNotEmpty {
				url := getUrlByMethod(methodSendMessage)
				url = url + "?chat_id=" + strconv.Itoa(update.Message.Chat.ID) + "&text=" + "Ищу дилера, который будет помогать в поиске сюжетов для картин и в продажах. https://vk.com/id75523881"
				getBodyByUrl(url)
			}
		}
		//sleep 5 seconds
		time.Sleep(5 * time.Second)
	}
}
	

func getUrlByMethod(methodName string) string {
	if methodName == "getUpdates" {
		methodName += "?offset=" + strconv.Itoa(offsetUpdates + 1)
	}
	return telegramBaseUrl + telegramToken + "/" + methodName
}

func getBodyByUrl(url string) []byte {
	response, err := http.Get(url)
	if err != nil {
		fmt.Println(err.Error())
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println(err.Error())
	}

	return body
}