package scrape

import (
	"database/sql"
	_ "github.com/ziutek/mymysql/godrv"
	"os"
	"strings"
	"time"
)

type DB struct {
	conn      *sql.DB
	stmtCache map[string]*sql.Stmt
}

func GetMySQLConnectionURI() string {
	if uri := os.Getenv("MYSQL_CONNECTION_URI"); uri != "" {
		return uri
	}
	return "scrapemonster//"
}

func OpenDatabase(uri string) (db *DB, err error) {
	var conn *sql.DB
	conn, err = sql.Open("mymysql", uri)
	if err == nil {
		db = &DB{conn: conn, stmtCache: make(map[string]*sql.Stmt)}
	}
	return
}

func (db *DB) StoreDeal(d *Deal) (err error) {
	var stmt *sql.Stmt
	stmt, err = db.getCachedStmt("insertDealDailySnapshot", insertDealDailySnapshotSQL)
	if err != nil {
		return
	}
	desc := trunc(d.Description, 500)
	cat := trunc(d.Category, 100)
	subcat := trunc(d.Subcategory, 100)
	locale := truncjoin(d.Locale, 200)
	_, err = stmt.Exec(d.SiteName, d.DealID,
		desc, cat, subcat, locale, d.OriginalPrice,
		d.DiscountPrice, d.NumSold, d.Expired, d.Adult,
		desc, cat, subcat, locale, d.OriginalPrice,
		d.DiscountPrice, d.NumSold, d.Expired, d.Adult)
	return
}

func (db *DB) StoreOption(o *Option) (err error) {
	var stmt *sql.Stmt
	stmt, err = db.getCachedStmt("insertOptionDailySnapshot", insertOptionDailySnapshotSQL)
	if err != nil {
		return
	}
	desc := trunc(&o.Description, 500)
	_, err = stmt.Exec(o.SiteName, o.DealID, o.OptionID,
		desc, o.Price, o.NumAvailable, o.NumSold,
		desc, o.Price, o.NumAvailable, o.NumSold)
	return
}

type DealDailySnapshot struct {
	Site          string
	DealID        int64
	Day           time.Time
	Description   *string
	Category      *string
	Subcategory   *string
	Locale        *string
	OriginalPrice *int
	DiscountPrice *int
	NumSold       *int
	IsExpired     bool
	IsAdult       bool
}

func (db *DB) GetDealDailySnapshots(day time.Time) (rs []*DealDailySnapshot, err error) {
	var rows *sql.Rows
	rows, err = db.conn.Query(selectDealDailySnapshotByDaySQL, day)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var r DealDailySnapshot
		err = rows.Scan(&r.Site, &r.DealID, &r.Day, &r.Description,
			&r.Category, &r.Subcategory, &r.Locale, &r.OriginalPrice,
			&r.DiscountPrice, &r.NumSold, &r.IsExpired, &r.IsAdult)
		if err != nil {
			return
		}
		rs = append(rs, &r)
	}
	return
}

type OptionDailySnapshot struct {
	Site         string
	DealID       int64
	OptionID     int64
	Day          time.Time
	Description  *string
	Price        *int
	NumAvailable *int
	NumSold      *int
}

func (db *DB) GetOptionDailySnapshots(day time.Time) (rs []*OptionDailySnapshot, err error) {
	var rows *sql.Rows
	rows, err = db.conn.Query(selectOptionDailySnapshotByDaySQL, day)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var r OptionDailySnapshot
		err = rows.Scan(&r.Site, &r.DealID, &r.OptionID, &r.Day,
			&r.Description, &r.Price, &r.NumAvailable, &r.NumSold)
		if err != nil {
			return
		}
		rs = append(rs, &r)
	}
	return
}

func (db *DB) getCachedStmt(name string, sql string) (stmt *sql.Stmt, err error) {
	if stmt = db.stmtCache[name]; stmt != nil {
		return
	}
	stmt, err = db.conn.Prepare(sql)
	if err != nil {
		return
	}
	db.stmtCache[name] = stmt
	return
}

func trunc(s *string, maxlen int) *string {
	if s == nil {
		return nil
	}
	t := strings.TrimSpace(*s)
	if len(t) > maxlen {
		t = t[:maxlen]
	}
	return &t
}

func truncjoin(a []string, maxlen int) *string {
	if len(a) == 0 {
		return nil
	}
	b := make([]string, 0, len(a))
	for _, s := range a {
		if t := strings.TrimSpace(s); t != "" {
			b = append(b, t)
		}
	}
	if len(b) == 0 {
		return nil
	}
	s := strings.Join(b, ", ")
	return trunc(&s, maxlen)
}

//
// SQL statements
//

const insertDealDailySnapshotSQL = `
    INSERT IGNORE INTO deal_daily_snapshot (
        site,
        deal_id,
        day,
        description,
        category,
        subcategory,
        locale,
        original_price,
        discount_price,
        num_sold,
        expired,
        adult,
        created,
        updated)
    VALUES (
        ?, /* site */
        ?, /* deal_id */
        CURRENT_DATE(), /* day */
        ?, /* description */
        ?, /* category */
        ?, /* subcategory */
        ?, /* locale */
        ?, /* original_price */
        ?, /* discount_price */
        ?, /* num_sold */
        ?, /* expired */
        ?, /* adult */
        NOW(), /* created */
        NOW()) /* updated */
    ON DUPLICATE KEY UPDATE
        updated = NOW(),
        description = ?,
        category = ?,
        subcategory = ?,
        locale = ?,
        original_price = ?,
        discount_price = ?,
        num_sold = ?,
        expired = ?,
        adult = ?`

const insertOptionDailySnapshotSQL = `
    INSERT IGNORE INTO option_daily_snapshot (
        site,
        deal_id,
        option_id,
        day,
        description,
        price,
        num_available,
        num_sold,
        created,
        updated)
    VALUES (
        ?, /* site */
        ?, /* deal_id */
        ?, /* option_id */
        CURRENT_DATE(), /* day */
        ?, /* description */
        ?, /* price */
        ?, /* num_available */
        ?, /* num_sold */
        NOW(), /* created */
        NOW()) /* updated */
    ON DUPLICATE KEY UPDATE
        updated = NOW(),
        description = ?,
        price = ?,
        num_available = ?,
        num_sold = ?`

const selectDealDailySnapshotByDaySQL = `
    SELECT site, deal_id, day, description, category, subcategory, locale,
        original_price, discount_price, num_sold, expired, adult
    FROM deal_daily_snapshot
    WHERE day = ?`

const selectOptionDailySnapshotByDaySQL = `
    SELECT site, deal_id, option_id, day, description,
        price, num_available, num_sold
    FROM option_daily_snapshot
    WHERE day = ?`
