package main

import (
	"database/sql" // Для работы с базой данных и SQL.
)

// ParcelStore - структура, реализующая хранилище данных.
// Содержит подключение к базе данных (db).
type ParcelStore struct {
	db *sql.DB
}

// NewParcelStore - конструктор для ParcelStore.
func NewParcelStore(db *sql.DB) ParcelStore {
	return ParcelStore{db: db}
}

// Add - добавляет новую посылку в таблицу 'parcel'.
// Возвращает присвоенный ID (Number) новой посылки.
func (s ParcelStore) Add(p Parcel) (int, error) {
	// Выполняем INSERT-запрос, используя именованные параметры.
	res, err := s.db.Exec("INSERT INTO parcel (client, status, address, created_at) VALUES (:client, :status, :address, :created_at)",
		sql.Named("client", p.Client),
		sql.Named("status", p.Status),
		sql.Named("address", p.Address),
		sql.Named("created_at", p.CreatedAt))
	if err != nil {
		return 0, err
	}

	// Получаем ID, автоматически сгенерированный базой данных.
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

// Get - извлекает информацию о посылке по ее уникальному номеру (Number).
func (s ParcelStore) Get(number int) (Parcel, error) {
	p := Parcel{}

	// Выполняем SELECT-запрос для получения одной строки.
	row := s.db.QueryRow("SELECT number, client, status, address, created_at FROM parcel WHERE number = :number",
		sql.Named("number", number))

	// Сканируем полученные значения в структуру Parcel.
	err := row.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
	if err != nil {
		// sql.ErrNoRows будет возвращен, если посылка не найдена.
		return p, err
	}

	return p, nil
}

// GetByClient - извлекает список всех посылок для конкретного клиента.
func (s ParcelStore) GetByClient(client int) ([]Parcel, error) {
	// Выполняем SELECT-запрос, который может вернуть несколько строк.
	rows, err := s.db.Query("SELECT number, client, status, address, created_at FROM parcel WHERE client = :client",
		sql.Named("client", client))
	if err != nil {
		return nil, err
	}
	defer rows.Close() // Обязательно закрываем курсор.

	var res []Parcel
	// Итерируемся по всем полученным строкам.
	for rows.Next() {
		p := Parcel{}

		// Сканируем текущую строку в структуру.
		err := rows.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
		if err != nil {
			return nil, err
		}

		res = append(res, p)
	}

	// Проверяем, не было ли ошибок при итерации по строкам.
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return res, nil
}

// SetStatus - обновляет статус посылки по ее номеру.
func (s ParcelStore) SetStatus(number int, status string) error {
	// Выполняем UPDATE-запрос.
	_, err := s.db.Exec("UPDATE parcel SET status = :status WHERE number = :number",
		sql.Named("status", status),
		sql.Named("number", number))

	return err
}

// SetAddress - обновляет адрес посылки по ее номеру.
// Обратите внимание: изменение адреса возможно ТОЛЬКО, если статус 'registered'.
func (s ParcelStore) SetAddress(number int, address string) error {
	_, err := s.db.Exec("UPDATE parcel SET address = :address WHERE number = :number AND status = :status",
		sql.Named("address", address),
		sql.Named("number", number),
		sql.Named("status", ParcelStatusRegistered))

	return err
}

// Delete - удаляет посылку по ее номеру.
// Обратите внимание: удаление возможно ТОЛЬКО, если статус 'registered'.
func (s ParcelStore) Delete(number int) error {
	_, err := s.db.Exec("DELETE FROM parcel WHERE number = :number AND status = :status",
		sql.Named("number", number),
		sql.Named("status", ParcelStatusRegistered)) // Используем константу статуса.

	return err
}
