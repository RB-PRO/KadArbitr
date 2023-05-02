package KadArbitr

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/playwright-community/playwright-go"
)

// Файл для парсинга каждой страницы судебного дела, [например]
//
// # Используется структура Card
//
// [пример]: https://kad.arbitr.ru/Card/72197155-c243-47d3-b328-2c421391754a
func (core *CoreReq) ParseCard(url string) (card Card, ErrorParse error) {

	// Переходим по ссылке с запроса
	if _, err := core.page.Goto(url); err != nil {
		return Card{}, err // could not create page
	}

	// Ждём загрузку определённой части страницы
	_, ErrorWait := core.page.WaitForSelector("dd[id=main-column]", playwright.PageWaitForSelectorOptions{Timeout: playwright.Float(20000)})
	if ErrorWait != nil {
		return Card{}, ErrorWait
	}

	// Тип дела
	if Selector, _ := core.page.QuerySelector(`dt[class^=b-iblock__header] span`); Selector != nil { // Если найден такой блок
		// Берём текстовое значение и проверяем его на ошибку
		if FindText, IsFindError := Selector.TextContent(); IsFindError == nil {
			card.Type = strings.TrimSpace(FindText)
		}
	}

	// Статус дела
	if Selector, _ := core.page.QuerySelector(`div[class=b-case-header-desc]`); Selector != nil { // Если найден такой блок
		// Берём текстовое значение и проверяем его на ошибку
		if FindText, IsFindError := Selector.TextContent(); IsFindError == nil {
			card.Status = strings.TrimSpace(FindText)
		}
	}

	// Следующее заседание
	if Selector, _ := core.page.QuerySelector(`div[class=b-instanceAdditional]`); Selector != nil { // Если найден такой блок
		// Берём текстовое значение и проверяем его на ошибку
		if FindText, IsFindError := Selector.TextContent(); IsFindError == nil {
			// FindText = strings.TrimSpace(FindText)
			FindText = strings.ReplaceAll(FindText, "Следующее заседание:", "")
			FindText = strings.TrimSpace(FindText)
			strs := strings.Split(FindText, ",")
			if len(strs) == 3 {
				// fmt.Println(">" + strs[0] + " " + strs[1] + "< - >" + strs[2] + "<")
				// Локация заседания
				card.Next.Location = strings.TrimSpace(strs[2])

				// Время сделающего заседания
				ParseTime, ErrorParse := time.Parse("02.01.2006 15:04", strings.TrimSpace(strs[0])+" "+strings.TrimSpace(strs[1]))
				if ErrorParse == nil {
					card.Next.Date = ParseTime
				}
			} else {
				fmt.Println("Не могу преобразовать данные для следующего заседания. Нужно добавить обработчик.\n" + FindText)
			}
		}
	}

	// Сперва пропарсим главные значения карточек
	MainsH, err := core.page.QuerySelectorAll(`div[class="b-chrono-item-header js-chrono-item-header page-break"] div[class="container container--live_translation"]`)
	if err == nil && len(MainsH) != 0 { // Если ненулевое к-во элементов
		// Выделяем память для карточек
		card.Slips = make([]struct {
			DataID string
			Main   HistoryMain
			Slave  []HistorySlave
		}, len(MainsH))
		for IndexMain, mainH := range MainsH { // Парсим каждую главную карточку
			// Название инстанции суда
			if Selector, _ := mainH.QuerySelector(`div[class=l-col] strong`); Selector != nil { // Если найден такой блок
				// Берём текстовое значение и проверяем его на ошибку
				if FindText, IsFindError := Selector.TextContent(); IsFindError == nil {
					FindText = strings.TrimSpace(FindText)
					card.Slips[IndexMain].Main.InstanceName = FindText
				}
			}

			// Дата
			if Selector, _ := mainH.QuerySelector(`div[class=l-col] span[class=b-reg-date]`); Selector != nil { // Если найден такой блок
				// Берём текстовое значение и проверяем его на ошибку
				if FindText, IsFindError := Selector.TextContent(); IsFindError == nil {
					FindText = strings.TrimSpace(FindText)
					ParseTime, ErrorParse := time.Parse("02.01.2006", FindText)
					if ErrorParse == nil {
						card.Slips[IndexMain].Main.Date = ParseTime
					}
				}
			}

			// Инстанции
			if Selector, _ := mainH.QuerySelector(`div[class=l-col] span[class=b-reg-incoming_num]`); Selector != nil { // Если найден такой блок
				// Берём текстовое значение и проверяем его на ошибку
				if FindText, IsFindError := Selector.TextContent(); IsFindError == nil {
					FindText = strings.TrimSpace(FindText)
					card.Slips[IndexMain].Main.Number = FindText
				}
			}

			// Отчет по датам публикаций
			if Selector, _ := mainH.QuerySelector(`div[class=r-col] h4 span[class=b-indentIcon] a`); Selector != nil { // Если найден такой блок
				// Берём текстовое значение и проверяем его на ошибку
				if FindText, EroorFind := Selector.GetAttribute("href"); EroorFind == nil {
					card.Slips[IndexMain].Main.UrlReport = FindText
				}
			}

			// Номер инстанции
			if Selector, _ := mainH.QuerySelector(`div[class=r-col] h4 span strong[class=b-case-instance-number]`); Selector != nil { // Если найден такой блок
				// Берём текстовое значение и проверяем его на ошибку
				if FindText, IsFindError := Selector.TextContent(); IsFindError == nil {
					FindText = strings.TrimSpace(FindText)
					card.Slips[IndexMain].Main.NumberInstance = FindText
				}
			}

			// Название суда
			if Selector, _ := mainH.QuerySelector(`div[class=r-col] h4 span span[class=instantion-name] a`); Selector != nil { // Если найден такой блок
				// Берём текстовое значение и проверяем его на ошибку
				if FindText, IsFindError := Selector.TextContent(); IsFindError == nil {
					FindText = strings.TrimSpace(FindText)
					card.Slips[IndexMain].Main.Cour = FindText
				}
			}

			// Ссылка на суд
			if Selector, _ := mainH.QuerySelector(`div[class=r-col] h4 span span[class=instantion-name] a`); Selector != nil { // Если найден такой блок
				// Берём текстовое значение и проверяем его на ошибку
				if FindText, EroorFind := Selector.GetAttribute("href"); EroorFind == nil {
					card.Slips[IndexMain].Main.UrlCour = FindText
				}
			}

			// Название файла
			if Selector, _ := mainH.QuerySelector(`div[class=r-col] h2[class=b-case-result] a`); Selector != nil { // Если найден такой блок
				// Берём текстовое значение и проверяем его на ошибку
				if FindText, IsFindError := Selector.TextContent(); IsFindError == nil {
					FindText = strings.TrimSpace(FindText)
					card.Slips[IndexMain].Main.FileName = FindText
				}
			}

			// Ссылка на файл
			if Selector, _ := mainH.QuerySelector(`div[class=r-col] h2[class=b-case-result] a`); Selector != nil { // Если найден такой блок
				// Берём текстовое значение и проверяем его на ошибку
				if FindText, EroorFind := Selector.GetAttribute("href"); EroorFind == nil {
					card.Slips[IndexMain].Main.FileLink = FindText
				}
			}
		}
	}

	// Теперь переходим к парсингу потомков

	// Нажимаем на все кнопки расширения
	MainsClick, err := core.page.QuerySelectorAll(`div[title="Нажмите, чтобы ознакомиться с полной хронологией дела."]`)
	if err == nil && len(MainsClick) != 0 { // Если ненулевое к-во элементов
		for _, value := range MainsClick {
			value.Click()
		}
	}

	time.Sleep(1 * time.Second)
	core.Screen("screens/Карточки2.jpg")

	// Цикл по всем строкам
	for i := 1; i <= len(MainsClick); i++ {
		MainsSlaves, err := core.page.QuerySelectorAll(`div[class="b-chrono-items-container js-chrono-items-container"]:nth-child(` + strconv.Itoa(i+1) + `) div[class=js-chrono-items-wrapper]`)
		if err == nil && len(MainsSlaves) != 0 { // Если ненулевое к-во элементов

			// Парсим записи в потомках
			MainsSlavesElement, ErrElem := core.page.QuerySelectorAll(`div[class^=b-chrono-item]`)
			if ErrElem == nil && len(MainsSlavesElement) != 0 {

				// Определяем массив потомков, в который и будем парсить
				// Далее приравняем данные в исходный массив элементов
				var slaves []HistorySlave

				// Счётчик всех элеменетов
				var cout int

				// Цикл по всем элеметам
				for _, Element := range MainsSlavesElement {
					var slave HistorySlave

					// Дата дела
					if Selector, _ := Element.QuerySelector(`div[class=l-col] p[class=case-date]`); Selector != nil { // Если найден такой блок
						// Берём текстовое значение и проверяем его на ошибку
						if FindText, IsFindError := Selector.TextContent(); IsFindError == nil {
							FindText = strings.TrimSpace(FindText)
							ParseTime, ErrorParse := time.Parse("02.01.2006", FindText)
							if ErrorParse == nil {
								slave.Date = ParseTime
							}
						}
					}
					// Тип дела
					if Selector, _ := Element.QuerySelector(`div[class=l-col] p[class=case-type]`); Selector != nil { // Если найден такой блок
						// Берём текстовое значение и проверяем его на ошибку
						if FindText, IsFindError := Selector.TextContent(); IsFindError == nil {
							FindText = strings.TrimSpace(FindText)
							fmt.Println(FindText)
							slave.Type = FindText
						}
					}
					// Информация о деле(там лежит сумма иска)
					if Selector, _ := Element.QuerySelector(`div[class=r-col] span[class=additional-info]`); Selector != nil { // Если найден такой блок
						// Берём текстовое значение и проверяем его на ошибку
						if FindText, IsFindError := Selector.TextContent(); IsFindError == nil {
							FindText = strings.TrimSpace(FindText)
							slave.Info = FindText
						}
					}

					// --- Инфомация о публикации ---
					// Ссылка на публикацию
					if Selector, _ := Element.QuerySelector(`div[class=r-col] p[class^=b-case-publish_info] a`); Selector != nil { // Если найден такой блок
						// Берём текстовое значение и проверяем его на ошибку
						if FindText, EroorFind := Selector.GetAttribute("href"); EroorFind == nil {
							slave.DatePost.URL = FindText
						}
					}
					// Дата публикации
					if Selector, _ := Element.QuerySelector(`div[class=r-col] p[class^=b-case-publish_info] a`); Selector != nil { // Если найден такой блок
						// Берём текстовое значение и проверяем его на ошибку
						if FindText, IsFindError := Selector.TextContent(); IsFindError == nil {
							FindText = strings.TrimSpace(FindText)
							FindText = strings.ReplaceAll(FindText, "Дата публикации:", "")
							FindText = strings.ReplaceAll(FindText, "г. ", "")
							FindText = strings.TrimSpace(FindText)
							FindText = FindText[:len(FindText)-4]
							FindText = strings.TrimSpace(FindText)
							//slave.DatePost.Time.
							slave.DatePost.Time, _ = time.Parse("02.01.2006 15:04:00", FindText)
						}
					}

					// --- Файл ---
					// Ссылка на файл
					if Selector, _ := Element.QuerySelector(`div[class=r-col] h2[class^=b-case-result] a[class^=b-case-result-text]`); Selector != nil { // Если найден такой блок
						// Берём текстовое значение и проверяем его на ошибку
						if FindText, EroorFind := Selector.GetAttribute("href"); EroorFind == nil {
							slave.Application.Link = FindText
						}
					}

					// Название файлаы
					if Selector, _ := Element.QuerySelector(`div[class=r-col] h2[class^=b-case-result] a[class^=b-case-result-text] span[class=js-judges-rollover]`); Selector != nil { // Если найден такой блок
						// Берём текстовое значение и проверяем его на ошибку
						if FindText, IsFindError := Selector.TextContent(); IsFindError == nil {
							FindText = strings.TrimSpace(FindText)
							slave.Application.Name = FindText
						}
					}

					// Суд
					if Selectors, _ := Element.QuerySelectorAll(`div[class=r-col] p[class^=case-subject] span`); Selectors != nil { // Если найден такой блок
						// Берём текстовое значение и проверяем его на ошибку
						// if FindText, IsFindError := Selector.TextContent(); IsFindError == nil {
						// 	FindText = strings.TrimSpace(FindText)
						// 	slave.Application.Name = FindText
						// }
						if len(Selectors) == 1 { // Если всего у нас один такой селектор
							if FindText, IsFindError := Selectors[0].TextContent(); IsFindError == nil {
								FindText = strings.TrimSpace(FindText)
								if strings.Count(FindText, ".") != 2 { // Если в строке нет двух точек, что говорит нам о том, что это не фамилия
									slave.JudgeOrCourt = FindText
								}
							}
						}
					}

					// Судебный состав, докладчик, судьи
					if Selectors, _ := Element.QuerySelectorAll(`div[class=r-col] h2[class^=b-case-result] span[class^=js-judges-rolloverHtml] p`); Selectors != nil { // Если найден такой блок
						for _, ValueSelector := range Selectors { // Цикл по всем параграм p, которые в теге strong содержат названия, которые и определяют, что это у нас. Буть то судьи, будь то Судебный состав
							if Selector, _ := ValueSelector.QuerySelector(`strong`); Selector != nil { // Если найден такой блок
								// Берём текстовое значение и проверяем его на ошибку
								if FindText, IsFindError := Selector.TextContent(); IsFindError == nil {
									FindText = strings.TrimSpace(FindText) //  FindText соожер информацию о том, что это такое.

									// Массив, который и будет заполняться данными из параграфа
									var FindesStrs []string

									//
									// Selector.
									// Костыль. Плохо. Так лучше не делать.
									html, _ := Selector.InnerHTML()
									html = strings.ReplaceAll(html, "<strong>"+FindText+"<strong>", "")
									strs := strings.Split(html, "<br>")
									for _, val := range strs {
										FindesStrs = append(FindesStrs, strings.TrimSpace(val))
									}

									// Заполнение полей по судьям
									switch FindText {
									case "Судебный состав:":
										slave.Application.JudicialComposition = FindesStrs
										break
									case "Судья-докладчик:":
										slave.Application.JudgeSpeaker = FindesStrs
										break
									case "Судьи:":
										slave.Application.Judges = FindesStrs
										break
									default:
										break
									}
								}
							}
						}
					}

					slaves = append(slaves, slave) // Добавляем элементы в массив
				}

				// Сохраняем данные
				card.Slips[cout].Slave = slaves

			}

		}
	}

	// b-chrono-items-container js-chrono-items-container

	return card, ErrorParse
}
