package main

import (
	"context"
	"fmt"
	"xhgm_price_tool/crawl"
	"xhgm_price_tool/good"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

func (a *App) Debug(args ...any) {
	fmt.Println(args...)
}

func (a *App) GetAllGoods() []*good.Good {
	return good.ItemArr
}

func (a *App) SetGoodPrice(name string, price int) {
	good.Name2Item[name].Price = price
}

func (a *App) GetAllBestProfit() (ps []good.Profit, err error) {
	ps, _, err = good.GetAllItemBestProfitByFocus(400, nil)
	return
}

func (a *App) RunCrawl() (err error) {
	_, err = crawl.Run(false)
	return
}

func (a *App) SavePrice() {
	good.SavePrice()
}
