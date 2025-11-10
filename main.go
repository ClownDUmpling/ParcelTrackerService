package main

import (
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite" // Импорт драйвера SQLite
)

// Константы для определения статусов посылки.
const (
	ParcelStatusRegistered = "registered" // Посылка зарегистрирована.
	ParcelStatusSent       = "sent"       // Посылка отправлена.
	ParcelStatusDelivered  = "delivered"  // Посылка доставлена.
)

// Parcel - структура, представляющая одну посылку.
type Parcel struct {
	Number    int    // Уникальный ID посылки.
	Client    int    // ID клиента.
	Status    string // Текущий статус.
	Address   string // Адрес доставки.
	CreatedAt string // Дата и время регистрации.
}

// ParcelService - сервис, реализующий бизнес-логику работы с посылками.
type ParcelService struct {
	store ParcelStore // Интерфейс к хранилищу данных (определен в parcel.go).
}

// NewParcelService - конструктор сервиса.
func NewParcelService(store ParcelStore) ParcelService {
	return ParcelService{store: store}
}

// Register - регистрирует новую посылку в системе.
func (s ParcelService) Register(client int, address string) (Parcel, error) {
	parcel := Parcel{
		Client:    client,
		Status:    ParcelStatusRegistered, // Устанавливаем начальный статус.
		Address:   address,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}

	// Добавляем посылку в хранилище и получаем ID.
	id, err := s.store.Add(parcel)
	if err != nil {
		return parcel, err
	}

	parcel.Number = id

	fmt.Printf("Новая посылка № %d на адрес %s от клиента с идентификатором %d зарегистрирована %s\n",
		parcel.Number, parcel.Address, parcel.Client, parcel.CreatedAt)

	return parcel, nil
}

// PrintClientParcels - получает и выводит в консоль все посылки клиента.
func (s ParcelService) PrintClientParcels(client int) error {
	// Получаем список посылок из хранилища.
	parcels, err := s.store.GetByClient(client)
	if err != nil {
		return err
	}

	fmt.Printf("Посылки клиента %d:\n", client)
	// Выводим данные каждой посылки.
	for _, parcel := range parcels {
		fmt.Printf("Посылка № %d на адрес %s от клиента с идентификатором %d зарегистрирована %s, статус %s\n",
			parcel.Number, parcel.Address, parcel.Client, parcel.CreatedAt, parcel.Status)
	}
	fmt.Println()

	return nil
}

// NextStatus - переводит посылку в следующий логический статус.
func (s ParcelService) NextStatus(number int) error {
	parcel, err := s.store.Get(number)
	if err != nil {
		return err
	}

	var nextStatus string
	// Логика перехода статусов.
	switch parcel.Status {
	case ParcelStatusRegistered:
		nextStatus = ParcelStatusSent
	case ParcelStatusSent:
		nextStatus = ParcelStatusDelivered
	case ParcelStatusDelivered:
		return nil // Нельзя изменить статус доставленной посылки.
	}

	fmt.Printf("У посылки № %d новый статус: %s\n", number, nextStatus)

	// Обновляем статус в БД.
	return s.store.SetStatus(number, nextStatus)
}

// ChangeAddress - изменяет адрес доставки.
func (s ParcelService) ChangeAddress(number int, address string) error {
	return s.store.SetAddress(number, address)
}

// Delete - удаляет посылку.
func (s ParcelService) Delete(number int) error {
	return s.store.Delete(number)
}

// main - основная точка входа. Демонстрирует работу сервиса.
func main() {
	// Открытие соединения с базой данных SQLite.
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer db.Close()

	// Инициализация хранилища и сервиса.
	store := NewParcelStore(db)
	service := NewParcelService(store)

	// --- Демонстрационный сценарий ---

	// 1. Регистрация первой посылки.
	client := 1
	address := "Псков, д. Пушкина, ул. Колотушкина, д. 5"
	p, err := service.Register(client, address)
	if err != nil {
		fmt.Println(err)
		return
	}

	// 2. Изменение адреса (допустимо, т.к. статус "registered").
	newAddress := "Саратов, д. Верхние Зори, ул. Козлова, д. 25"
	err = service.ChangeAddress(p.Number, newAddress)
	if err != nil {
		fmt.Println(err)
		return
	}

	// 3. Изменение статуса на "sent".
	err = service.NextStatus(p.Number)
	if err != nil {
		fmt.Println(err)
		return
	}

	// 4. Вывод посылок клиента.
	err = service.PrintClientParcels(client)
	if err != nil {
		fmt.Println(err)
		return
	}

	// 5. Попытка удаления отправленной посылки (должна завершиться ошибкой).
	err = service.Delete(p.Number)
	if err != nil {
		fmt.Println(err) // Вывод ожидаемой ошибки.
		return
	}

	// 6. Повторный вывод посылок (первая посылка не удалилась).
	err = service.PrintClientParcels(client)
	if err != nil {
		fmt.Println(err)
		return
	}

	// 7. Регистрация новой посылки.
	p, err = service.Register(client, address)
	if err != nil {
		fmt.Println(err)
		return
	}

	// 8. Удаление новой посылки (должно пройти успешно, т.к. статус "registered").
	err = service.Delete(p.Number)
	if err != nil {
		fmt.Println(err)
		return
	}

	// 9. Финальный вывод посылок (вторая посылка должна отсутствовать).
	err = service.PrintClientParcels(client)
	if err != nil {
		fmt.Println(err)
		return
	}
}
