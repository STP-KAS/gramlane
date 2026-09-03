package server

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log"
	"net/http"
	"strconv"
	"strings"

	"gramlane/internal/agent"
	"gramlane/internal/desk"
	"gramlane/internal/feedback"
	"gramlane/internal/framing"
	"gramlane/internal/genesis"
	"gramlane/internal/jobs"
	"gramlane/internal/names"
	"gramlane/internal/pos"
	"gramlane/internal/post"
	"gramlane/internal/quote"
	"gramlane/internal/seq"
	"gramlane/internal/wallets"
	"gramlane/web"

	qrcode "github.com/skip2/go-qrcode"
)

type Server struct {
	Addr string
	T    *template.Template
}

type page struct {
	Title     string
	Active    string
	Query     string
	Error     string
	Jobs      []jobs.Job
	Job       *jobs.Job
	Quote     *quote.Quote
	Run       *jobs.Receipt
	Wallets   []wallets.Wallet
	Framing   *framing.View
	PayTo     string
	Genesis   *genesis.Plan
	Seq       *seq.Ledger
	Stamps    []post.Stamp
	Site      *names.Page
	Card      *agent.Card
	Reply     *agent.Reply
	HasGrok   bool
	Conv      *quote.Convert
	Fits      []jobs.Fit
	Aside     *quote.Aside
	Invoice   *pos.Invoice
	Invoices  []pos.Invoice
	Merchants []pos.Merchant
	PayURL    string
	Sign      *pos.Sign
}

func New(addr string) (*Server, error) {
	t, err := template.ParseFS(web.Templates, "templates/*.html")
	if err != nil {
		return nil, err
	}
	return &Server{Addr: addr, T: t}, nil
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	static, err := fs.Sub(web.Static, "static")
	if err != nil {
		log.Fatal(err)
	}
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(static))))
	mux.HandleFunc("/", s.home)
	mux.HandleFunc("/desk", s.desk)
	mux.HandleFunc("/work", s.desk)
	mux.HandleFunc("/till", s.posPage)
	mux.HandleFunc("/convert", s.convertPage)
	mux.HandleFunc("/api/convert", s.apiConvert)
	mux.HandleFunc("/spend", s.spendPage)
	mux.HandleFunc("/stablegram", s.spendPage)
	mux.HandleFunc("/pos", s.posPage)
	mux.HandleFunc("/pay/", s.payPage)
	mux.HandleFunc("/counter/", s.counterPage)
	mux.HandleFunc("/qr/", s.qrPNG)
	mux.HandleFunc("/go", s.goPay)
	mux.HandleFunc("/api/pos", s.apiPos)
	mux.HandleFunc("/job/", s.job)
	mux.HandleFunc("/run", s.runPage)
	mux.HandleFunc("/honest", s.honest)
	mux.HandleFunc("/docs", s.docs)
	mux.HandleFunc("/guide", s.guidePage)
	mux.HandleFunc("/steps", s.guidePage)
	mux.HandleFunc("/genesis", s.genesisPage)
	mux.HandleFunc("/api/genesis", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, map[string]any{"ok": true, "data": genesis.Live(), "seq": seq.Snap(), "payTo": desk.PayTo()})
	})
	mux.HandleFunc("/api/seq", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, map[string]any{"ok": true, "data": seq.Snap()})
	})
	mux.HandleFunc("/api/genesis/artifact", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "contracts/v1/WorkCredit-live.json")
	})
	mux.HandleFunc("/wallets", func(w http.ResponseWriter, r *http.Request) {
		s.render(w, "wallets.html", page{Title: "Wallets · Gramlane", Active: "wallets", Wallets: wallets.All()})
	})
	mux.HandleFunc("/api/wallets", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, map[string]any{"ok": true, "inject": []string{"kasware", "kastle"}, "ledger": "https://kasvault.io", "data": wallets.All()})
	})
	mux.HandleFunc("/safety", func(w http.ResponseWriter, r *http.Request) {
		s.render(w, "safety.html", page{Title: "Safety · Gramlane", Active: "safety"})
	})
	mux.HandleFunc("/vision", s.whyPage)
	mux.HandleFunc("/explain", s.whyPage)
	mux.HandleFunc("/idea", s.whyPage)
	mux.HandleFunc("/why", s.whyPage)
	mux.HandleFunc("/234", s.framingPage)
	mux.HandleFunc("/framing", s.framingPage)
	mux.HandleFunc("/vault", s.vaultPage)
	mux.HandleFunc("/postage", s.postagePage)
	mux.HandleFunc("/post", s.postagePage)
	mux.HandleFunc("/agent", s.agentPage)
	mux.HandleFunc("/api/agent", s.apiAgent)
	mux.HandleFunc("/site/", s.siteName)
	mux.HandleFunc("/site", s.sitePage)
	mux.HandleFunc("/api/framing", func(w http.ResponseWriter, r *http.Request) {
		v := framing.Demo()
		if hx := strings.TrimSpace(r.URL.Query().Get("hex")); hx != "" {
			got, err := framing.DecodeHex(hx)
			if err != nil {
				writeJSON(w, 400, map[string]any{"ok": false, "error": err.Error()})
				return
			}
			v.Custom = &got
		}
		writeJSON(w, 200, map[string]any{"ok": true, "data": v})
	})
	mux.HandleFunc("/feedback", s.feedbackPage)
	mux.HandleFunc("/api/feedback", s.apiFeedback)
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		led := seq.Snap()
		writeJSON(w, 200, map[string]any{
			"ok": true, "dapp": "gramlane", "layer": "kaspa-l1",
			"unit": "gram", "l2": false, "stablecoin": false,
			"vision":         "stable work price on L1, not a synthetic dollar",
			"products":       []string{"spend", "pos", "vault", "postage", "agent", "site", "convert"},
			"sompiPerGram":   quote.SompiPerGram,
			"grok":           agent.HasKey(),
			"gramsRemaining": led.Remaining, "credits": led.Credits,
			"voucherOnChain": led.OnChain, "saleTx": led.SaleTx,
			"voucherTx": led.VoucherTx, "p2sh": led.P2SH,
		})
	})
	mux.HandleFunc("/api/jobs", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, map[string]any{"ok": true, "data": jobs.Catalog})
	})
	mux.HandleFunc("/api/quote", s.apiQuote)
	mux.HandleFunc("/api/run", s.apiRun)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/static/") {
			log.Printf("%s %s", r.Method, r.URL.Path)
		}
		mux.ServeHTTP(w, r)
	})
}

func (s *Server) render(w http.ResponseWriter, name string, p page) {
	if p.Seq == nil {
		snap := seq.Snap()
		p.Seq = &snap
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.T.ExecuteTemplate(w, name, p); err != nil {
		log.Println("template", name, err)
		http.Error(w, err.Error(), 500)
	}
}

func (s *Server) framingPage(w http.ResponseWriter, r *http.Request) {
	v := framing.Demo()
	p := page{Title: "#234 · Gramlane", Active: "234", Framing: &v}
	if hx := strings.TrimSpace(r.FormValue("hex")); hx != "" {
		p.Query = hx
		got, err := framing.DecodeHex(hx)
		if err != nil {
			p.Error = err.Error()
		} else {
			v.Custom = &got
			p.Framing = &v
		}
	}
	s.render(w, "framing.html", p)
}

func (s *Server) home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	s.render(w, "home.html", page{Title: "Gramlane", Active: "home"})
}

func (s *Server) whyPage(w http.ResponseWriter, r *http.Request) {
	s.render(w, "why.html", page{Title: "Why · Gramlane", Active: "why"})
}

func (s *Server) desk(w http.ResponseWriter, r *http.Request) {
	s.render(w, "desk.html", page{Title: "Work · Gramlane", Active: "work", Jobs: jobs.Catalog})
}

func convertFromRequest(r *http.Request) (quote.Convert, error) {
	kind := r.FormValue("unit")
	amount := strings.TrimSpace(r.FormValue("amount"))
	if amount == "" {
		amount = strings.TrimSpace(r.URL.Query().Get("kas"))
		if amount != "" {
			kind = "kas"
		}
	}
	if amount == "" {
		amount = strings.TrimSpace(r.URL.Query().Get("grams"))
		if amount != "" {
			kind = "grams"
		}
	}
	if amount == "" {
		amount = strings.TrimSpace(r.URL.Query().Get("sompi"))
		if amount != "" {
			kind = "sompi"
		}
	}
	if r.FormValue("from") == "remaining" || r.URL.Query().Get("from") == "remaining" {
		return quote.FromGrams(seq.Snap().Remaining)
	}
	if amount == "" {
		return quote.Convert{}, fmt.Errorf("amount")
	}
	if kind == "" {
		kind = "kas"
	}
	return quote.Parse(kind, amount)
}

func (s *Server) convertPage(w http.ResponseWriter, r *http.Request) {
	p := page{Title: "Convert · Gramlane", Active: "convert", Query: r.FormValue("amount")}
	has := r.FormValue("amount") != "" || r.FormValue("from") == "remaining" ||
		r.URL.Query().Get("kas") != "" || r.URL.Query().Get("grams") != "" ||
		r.URL.Query().Get("sompi") != "" || r.URL.Query().Get("from") == "remaining"
	if has {
		c, err := convertFromRequest(r)
		if err != nil {
			p.Error = err.Error()
		} else {
			p.Conv = &c
			p.Fits = jobs.Fits(c.Grams)
		}
	}
	s.render(w, "convert.html", p)
}

func (s *Server) apiConvert(w http.ResponseWriter, r *http.Request) {
	c, err := convertFromRequest(r)
	if err != nil {
		writeJSON(w, 400, map[string]any{"ok": false, "error": err.Error(), "hint": "kas=0.5 or grams=8000 or sompi=50000000 or from=remaining"})
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true, "convert": c, "fits": jobs.Fits(c.Grams), "seq": seq.Snap()})
}

func origin(r *http.Request) string {
	host := r.Host
	if host == "" {
		host = "127.0.0.1:8081"
	}
	return "http://" + host
}

func (s *Server) spendPage(w http.ResponseWriter, r *http.Request) {
	total := strings.TrimSpace(r.FormValue("total"))
	if total == "" {
		total = "1"
	}
	pct := uint64(50)
	if n, err := strconv.ParseUint(r.FormValue("pct"), 10, 64); err == nil {
		pct = n
	}
	p := page{Title: "Stablegram · Gramlane", Active: "stablegram", Query: total, PayTo: desk.PayTo()}
	a, err := quote.SetAside(total, pct)
	if err != nil {
		p.Error = err.Error()
	} else {
		p.Aside = &a
	}
	if r.Method == http.MethodPost {
		tx := strings.TrimSpace(r.FormValue("payment"))
		if tx != "" && p.Aside != nil {
			grams := p.Aside.PaySompi / quote.SompiPerGram
			if _, err := seq.MintFromTx(tx, grams, p.Aside.PaySompi); err != nil {
				p.Error = err.Error()
			}
		}
	}
	s.render(w, "spend.html", p)
}

func (s *Server) posPage(w http.ResponseWriter, r *http.Request) {
	p := page{Title: "Till · Gramlane", Active: "till", PayTo: desk.PayTo()}
	if r.Method == http.MethodPost {
		grams := uint64(0)
		if n, err := strconv.ParseUint(strings.ReplaceAll(r.FormValue("grams"), ",", ""), 10, 64); err == nil {
			grams = n
		}
		if grams == 0 && r.FormValue("kas") != "" {
			if c, err := quote.FromKAS(r.FormValue("kas")); err == nil {
				grams = c.Grams
			}
		}
		fiat := strings.TrimSpace(r.FormValue("fiat"))
		ccy := strings.TrimSpace(r.FormValue("ccy"))
		rate := strings.TrimSpace(r.FormValue("rate"))
		var bill quote.FiatBill
		if grams == 0 && fiat != "" {
			b, err := quote.FromFiat(fiat, ccy, rate)
			if err != nil {
				p.Error = err.Error()
			} else {
				bill = b
				grams = b.Grams
				pos.SaveSign(pos.Sign{Ccy: b.Ccy, KasInFiat: b.KasInFiat})
			}
		} else if ccy != "" && rate != "" {
			pos.SaveSign(pos.Sign{Ccy: ccy, KasInFiat: rate})
		}
		if p.Error == "" {
			inv, err := pos.Create(r.FormValue("item"), grams, r.FormValue("merchant"), r.FormValue("payTo"), r.FormValue("place"))
			if err != nil {
				p.Error = err.Error()
			} else {
				if bill.Grams > 0 {
					inv = pos.SetShelf(inv.ID, bill.Amount, bill.Ccy, bill.KasInFiat, bill.Label)
				}
				p.Invoice = inv
				p.PayURL = pos.PayURL(origin(r), inv.ID)
			}
		}
	}
	p.Invoices = pos.List()
	p.Merchants = pos.Merchants()
	sg := pos.LoadSign()
	p.Sign = &sg
	s.render(w, "pos.html", p)
}

func (s *Server) payPage(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/pay/")
	inv, ok := pos.Get(id)
	if !ok {
		http.NotFound(w, r)
		return
	}
	p := page{Title: "Pay · Gramlane", Active: "till", Invoice: inv, PayURL: pos.PayURL(origin(r), inv.ID), PayTo: inv.PayTo}
	if r.Method == http.MethodPost && inv.Status != "paid" {
		paid, err := pos.Pay(inv.ID, r.FormValue("payer"))
		if err != nil {
			p.Error = err.Error()
		}
		if paid != nil {
			p.Invoice = paid
		}
	}
	s.render(w, "pay.html", p)
}

func (s *Server) goPay(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.FormValue("id"))
	id = strings.TrimPrefix(id, "/pay/")
	if id == "" {
		http.Redirect(w, r, "/till", http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/pay/"+id, http.StatusSeeOther)
}

func (s *Server) counterPage(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/counter/")
	inv, ok := pos.Get(id)
	if !ok {
		http.NotFound(w, r)
		return
	}
	s.render(w, "counter.html", page{
		Title: inv.Item + " · till", Active: "till", Invoice: inv,
		PayURL: pos.PayURL(origin(r), inv.ID),
	})
}

func (s *Server) qrPNG(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/qr/")
	id = strings.TrimSuffix(id, ".png")
	inv, ok := pos.Get(id)
	if !ok {
		http.NotFound(w, r)
		return
	}
	u := pos.PayURL(origin(r), inv.ID)
	png, err := qrcode.Encode(u, qrcode.Medium, 256)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write(png)
}

func (s *Server) apiPos(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		if id := strings.TrimSpace(r.URL.Query().Get("id")); id != "" {
			inv, ok := pos.Get(id)
			if !ok {
				writeJSON(w, 404, map[string]any{"ok": false, "error": "unknown invoice"})
				return
			}
			writeJSON(w, 200, map[string]any{"ok": true, "invoice": inv, "seq": seq.Snap()})
			return
		}
		writeJSON(w, 200, map[string]any{"ok": true, "invoices": pos.List(), "merchants": pos.Merchants(), "seq": seq.Snap()})
		return
	}
	if r.Method == http.MethodPost {
		var req struct {
			Item, Merchant, PayTo, Place, ID, Payer string
			Grams                                   uint64
		}
		body, _ := io.ReadAll(io.LimitReader(r.Body, 1<<16))
		_ = json.Unmarshal(body, &req)
		if req.ID != "" {
			inv, err := pos.Pay(req.ID, req.Payer)
			if err != nil {
				writeJSON(w, 400, map[string]any{"ok": false, "error": err.Error(), "invoice": inv})
				return
			}
			writeJSON(w, 200, map[string]any{"ok": true, "invoice": inv, "seq": seq.Snap()})
			return
		}
		inv, err := pos.Create(req.Item, req.Grams, req.Merchant, req.PayTo, req.Place)
		if err != nil {
			writeJSON(w, 400, map[string]any{"ok": false, "error": err.Error()})
			return
		}
		writeJSON(w, 200, map[string]any{"ok": true, "invoice": inv, "pay": pos.PayURL(origin(r), inv.ID)})
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true, "invoices": pos.List(), "merchants": pos.Merchants(), "seq": seq.Snap()})
}

func (s *Server) job(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/job/")
	j, ok := jobs.Get(id)
	if !ok {
		http.NotFound(w, r)
		return
	}
	q, err := jobs.QuoteJob(j)
	p := page{Title: j.Name + " · Gramlane", Active: "work", Job: &j, Query: r.URL.Query().Get("q"), PayTo: desk.PayTo()}
	if err != nil {
		p.Error = err.Error()
	} else {
		p.Quote = &q
	}
	s.render(w, "job.html", p)
}

func (s *Server) honest(w http.ResponseWriter, r *http.Request) {
	s.render(w, "honest.html", page{Title: "Claims · Gramlane", Active: "honest"})
}

func (s *Server) docs(w http.ResponseWriter, r *http.Request) {
	s.render(w, "docs.html", page{Title: "Docs · Gramlane", Active: "docs"})
}

func (s *Server) guidePage(w http.ResponseWriter, r *http.Request) {
	s.render(w, "guide.html", page{Title: "Guide · Gramlane", Active: "guide"})
}

func (s *Server) genesisPage(w http.ResponseWriter, r *http.Request) {
	p := genesis.Live()
	s.render(w, "genesis.html", page{Title: "Genesis · Gramlane", Active: "genesis", Genesis: &p, PayTo: desk.PayTo()})
}

func (s *Server) burnJob(id, q, payer string) (jobs.Job, jobs.Receipt, error) {
	j, ok := jobs.Get(id)
	if !ok {
		return jobs.Job{}, jobs.Receipt{}, fmt.Errorf("unknown job")
	}
	paid := settlePaid(j, "grams")
	if paid == "" {
		return j, jobs.Receipt{}, fmt.Errorf("prepaid grams are spent")
	}
	rec, err := jobs.RunAs(j, q, paid, payer, "")
	if err != nil {
		return j, rec, err
	}
	applyBurn(&rec, j, paid)
	return j, rec, nil
}

func (s *Server) vaultPage(w http.ResponseWriter, r *http.Request) {
	v := framing.Demo()
	j, _ := jobs.Get("vault")
	p := page{Title: "Vault bump · Gramlane", Active: "work", Framing: &v, Job: &j, PayTo: desk.PayTo(), Query: r.FormValue("hex")}
	if hx := strings.TrimSpace(p.Query); hx != "" {
		got, err := framing.DecodeHex(hx)
		if err != nil {
			p.Error = err.Error()
		} else {
			v.Custom = &got
			p.Framing = &v
		}
	}
	if r.Method == http.MethodPost {
		_, rec, err := s.burnJob("vault", p.Query, r.FormValue("payer"))
		p.Run = &rec
		if err != nil {
			p.Error = err.Error()
		}
	}
	s.render(w, "vault.html", p)
}

func (s *Server) postagePage(w http.ResponseWriter, r *http.Request) {
	j, _ := jobs.Get("postage")
	p := page{Title: "Postage · Gramlane", Active: "work", Job: &j, PayTo: desk.PayTo(), Query: r.FormValue("q")}
	if r.Method == http.MethodPost {
		_, rec, err := s.burnJob("postage", p.Query, r.FormValue("payer"))
		p.Run = &rec
		if err != nil {
			p.Error = err.Error()
		}
	}
	p.Stamps = post.List()
	s.render(w, "postage.html", p)
}

func (s *Server) agentPage(w http.ResponseWriter, r *http.Request) {
	j, _ := jobs.Get("agent")
	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		name = strings.TrimSpace(r.URL.Query().Get("q"))
	}
	prompt := strings.TrimSpace(r.FormValue("prompt"))
	p := page{Title: "Agent · Gramlane", Active: "work", Job: &j, PayTo: desk.PayTo(), Query: name, HasGrok: agent.HasKey()}
	if name != "" && r.Method == http.MethodGet {
		if c, err := agent.CardFor(name); err == nil {
			p.Card = c
		} else {
			p.Error = err.Error()
		}
	}
	if r.Method == http.MethodPost {
		q := name
		if prompt != "" {
			if name != "" {
				q = name + " | " + prompt
			} else {
				q = prompt
			}
		}
		_, rec, err := s.burnJob("agent", q, r.FormValue("payer"))
		p.Run = &rec
		if err != nil {
			p.Error = err.Error()
		}
		if name != "" {
			if c, err := agent.CardFor(name); err == nil {
				p.Card = c
			}
		}
		if rec.Output != "" && prompt != "" {
			var rep agent.Reply
			if json.Unmarshal([]byte(rec.Output), &rep) == nil {
				p.Reply = &rep
			}
		}
	}
	s.render(w, "agent.html", p)
}

func (s *Server) apiAgent(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimSpace(r.URL.Query().Get("name"))
	if name == "" {
		name = "kns.kas"
	}
	if r.Method == http.MethodGet && r.URL.Query().Get("ask") == "" {
		c, err := agent.CardFor(name)
		if err != nil {
			writeJSON(w, 502, map[string]any{"ok": false, "error": err.Error()})
			return
		}
		writeJSON(w, 200, map[string]any{"ok": true, "card": c, "grok": agent.HasKey()})
		return
	}
	j, rec, err := s.burnJob("agent", name+" | "+r.URL.Query().Get("ask"), "")
	if err != nil {
		writeJSON(w, 402, map[string]any{"ok": false, "error": err.Error(), "job": j, "receipt": rec})
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true, "receipt": rec, "seq": seq.Snap(), "grok": agent.HasKey()})
}

func (s *Server) sitePage(w http.ResponseWriter, r *http.Request) {
	j, _ := jobs.Get("site")
	q := strings.TrimSpace(r.FormValue("q"))
	if q == "" {
		q = strings.TrimSpace(r.URL.Query().Get("q"))
	}
	p := page{Title: "Site builder · Gramlane", Active: "work", Job: &j, PayTo: desk.PayTo(), Query: q}
	if r.Method == http.MethodPost {
		if q == "" {
			q = "kns.kas"
		}
		_, rec, err := s.burnJob("site", q, r.FormValue("payer"))
		p.Run = &rec
		if err != nil {
			p.Error = err.Error()
		} else if site, lerr := names.Lookup(q); lerr == nil {
			p.Site = site
		}
	}
	s.render(w, "site.html", p)
}

func (s *Server) siteName(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/site/")
	if name == "" {
		s.sitePage(w, r)
		return
	}
	j, _ := jobs.Get("site")
	site, err := names.Lookup(name)
	p := page{Title: names.Normalize(name) + " · Gramlane", Active: "work", Job: &j, Site: site, Query: names.Normalize(name)}
	if err != nil {
		p.Error = err.Error()
	}
	s.render(w, "domain.html", p)
}

func (s *Server) runPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		http.Redirect(w, r, "/desk", http.StatusSeeOther)
		return
	}
	id := r.FormValue("job")
	q := r.FormValue("q")
	j, ok := jobs.Get(id)
	if !ok {
		s.render(w, "job.html", page{Title: "Run", Active: "work", Error: "unknown job"})
		return
	}
	paid := strings.TrimSpace(r.FormValue("payment"))
	payer := strings.TrimSpace(r.FormValue("payer"))
	wallet := strings.TrimSpace(r.FormValue("wallet"))
	if paid == "" && payer != "" {
		paid = wallet + ":" + payer
	}
	paid = settlePaid(j, paid)
	if paid == "" {
		qq, _ := jobs.QuoteJob(j)
		s.render(w, "job.html", page{
			Title: j.Name, Active: "work", Job: &j, Query: q, Quote: &qq,
			Error: "Prepaid grams are spent. Pay KAS fallback or paste a receipt id.",
		})
		return
	}
	rec, err := jobs.RunAs(j, q, paid, payer, wallet)
	p := page{Title: "Receipt · Gramlane", Active: "work", Job: &j, Query: q, Run: &rec}
	if err != nil {
		p.Error = err.Error()
		s.render(w, "run.html", p)
		return
	}
	applyBurn(&rec, j, paid)
	p.Run = &rec
	s.render(w, "run.html", p)
}

func (s *Server) apiQuote(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("job")
	j, ok := jobs.Get(id)
	if !ok {
		writeJSON(w, 404, map[string]any{"ok": false, "error": "unknown job"})
		return
	}
	q, err := jobs.QuoteJob(j)
	if err != nil {
		writeJSON(w, 400, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	led := seq.Snap()
	writeJSON(w, 200, map[string]any{"ok": true, "job": j, "quote": q, "seq": led, "prepaid": led.Remaining >= j.Grams})
}

func (s *Server) apiRun(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("job")
	if id == "" && r.Method == http.MethodPost {
		body, _ := io.ReadAll(io.LimitReader(r.Body, 1<<16))
		var req struct{ Job, Q, Payment string }
		_ = json.Unmarshal(body, &req)
		id = req.Job
		if r.URL.Query().Get("q") == "" {
			r.URL.RawQuery = "job=" + req.Job + "&q=" + req.Q
		}
		if req.Payment != "" {
			r.Header.Set("X-Work-Credit", req.Payment)
		}
	}
	j, ok := jobs.Get(id)
	if !ok {
		writeJSON(w, 404, map[string]any{"ok": false, "error": "unknown job"})
		return
	}
	qname := r.URL.Query().Get("q")
	paid := strings.TrimSpace(r.Header.Get("X-Work-Credit"))
	if paid == "" {
		paid = strings.TrimSpace(r.Header.Get("X-Kaspa-Payment"))
	}
	payer := strings.TrimSpace(r.Header.Get("X-Kaspa-Payer"))
	wallet := strings.TrimSpace(r.Header.Get("X-Kaspa-Wallet"))
	if paid == "" && payer != "" {
		paid = wallet + ":" + payer
	}
	paid = settlePaid(j, paid)
	if paid == "" {
		qq, _ := jobs.QuoteJob(j)
		led := seq.Snap()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusPaymentRequired)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error":          "Payment Required",
			"note":           "Prepaid grams are spent. Burn WorkCredit grams or pay KAS fallback. Not USD. No L2. Header X-Work-Credit or X-Kaspa-Payment.",
			"job":            j,
			"quote":          qq,
			"gramsRemaining": led.Remaining,
			"accepts": []map[string]any{
				{"scheme": "kaspa-work-credit", "asset": "GRAM", "grams": j.Grams, "lane": j.Lane},
				{"scheme": "kaspa", "asset": "KAS", "sompi": qq.Sompi},
			},
		})
		return
	}
	rec, err := jobs.RunAs(j, qname, paid, payer, wallet)
	if err != nil {
		writeJSON(w, 502, map[string]any{"ok": false, "error": err.Error(), "receipt": rec})
		return
	}
	applyBurn(&rec, j, paid)
	writeJSON(w, 200, map[string]any{"ok": true, "receipt": rec, "seq": seq.Snap()})
}

func (s *Server) feedbackPage(w http.ResponseWriter, r *http.Request) {
	p := page{Title: "Feedback · Gramlane", Active: "feedback"}
	if r.Method == http.MethodPost {
		n, err := feedback.Save("gramlane", r.FormValue("text"), r.FormValue("contact"))
		if err != nil {
			p.Error = err.Error()
		} else {
			p.Query = n.ID
		}
	}
	s.render(w, "feedback.html", p)
}

func (s *Server) apiFeedback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, 405, map[string]any{"ok": false, "error": "POST"})
		return
	}
	n, err := feedback.Save("gramlane", r.FormValue("text"), r.FormValue("contact"))
	if err != nil {
		body, _ := io.ReadAll(io.LimitReader(r.Body, 1<<16))
		var req struct{ Text, Contact string }
		_ = json.Unmarshal(body, &req)
		n, err = feedback.Save("gramlane", req.Text, req.Contact)
	}
	if err != nil {
		writeJSON(w, 400, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true, "id": n.ID, "dir": feedback.Dir(), "note": "This site never DMs you."})
}

func settlePaid(j jobs.Job, paid string) string {
	paid = strings.TrimSpace(paid)
	if seq.CanBurn(j.Grams) {
		if paid == "" || seq.Accepts(paid) {
			if paid == "" {
				return "grams"
			}
			return paid
		}
		return paid
	}
	if seq.Accepts(paid) {
		return ""
	}
	return paid
}

func applyBurn(rec *jobs.Receipt, j jobs.Job, paid string) {
	if !seq.Accepts(paid) {
		return
	}
	led, err := seq.BurnGrams(j.ID, j.Grams, paid)
	if err != nil {
		if rec.TxNote != "" {
			rec.TxNote += " — "
		}
		rec.TxNote += "ledger: " + err.Error()
		return
	}
	rec.Settlement = "prepaid-grams"
	rec.Remaining = led.Remaining
	rec.Note = "Sequencer inventory from the 0.5 KAS sale + P2SH UTXO. This burn is operator accounting until consume() spends that UTXO."
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
