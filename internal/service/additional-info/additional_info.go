// Пакет addinfo нужен для формирования стуктуры AddInfo, которая содержит
// поля дополнительной информации, необходимые для работы сервера
package addinfo

// AddInfo - структура дополнительных полей
type AddInfo struct {
	RemoteAddr string // ip-адрес агента
}
