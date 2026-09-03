package server

import (
	"encoding/json"
	"html/template"
	"io"
	"io/fs"
	"log"
	"net/http"
	"strings"

	"gramlane/internal/feedback"
	"gramlane/internal/framing"
	"gramlane/internal/jobs"
	"gramlane/internal/quote"
	"gramlane/internal/wallets"
	"gramlane/web"
)

type Server struct {
	Addr string
	T    *template.Template
}

type page struct {
	Title   string
	Active  string
	Query   string
	Error   string
	Jobs    []jobs.Job
	Job     *jobs.Job
	Quote   *quote.Quote
	Run     *jobs.Receipt
	Wallets []wallets.Wallet
	Framing *framing.View
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
	mux.HandleFunc("/job/", s.job)
	mux.HandleFunc("/run", s.runPage)
	mux.HandleFunc("/honest", s.honest)
	mux.HandleFunc("/docs", s.docs)
	mux.HandleFunc("/guide", s.guidePage)
	mux.HandleFunc("/steps", s.guidePage)
	mux.HandleFunc("/wallets", func(w http.ResponseWriter, r *http.Request) {
		s.render(w, "wallets.html", page{Title: "Wallets · Gramlane", Active: "wallets", Wallets: wallets.All()})
	})
	mux.HandleFunc("/api/wallets", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, map[string]any{"ok": true, "inject": []string{"kasware", "kastle"}, "ledger": "https://kasvault.io", "data": wallets.All()})
	})
	mux.HandleFunc("/safety", func(w http.ResponseWriter, r *http.Request) {
		s.render(w, "safety.html", page{Title: "Safety · Gramlane", Active: "safety"})
	})
	mux.HandleFunc("/idea", func(w http.ResponseWriter, r *http.Request) {
		s.render(w, "idea.html", page{Title: "Idea · Gramlane", Active: "idea"})
	})
	mux.HandleFunc("/explain", func(w http.ResponseWriter, r *http.Request) {
		s.render(w, "idea.html", page{Title: "Idea · Gramlane", Active: "idea"})
	})
	mux.HandleFunc("/why", func(w http.ResponseWriter, r *http.Request) {
		s.render(w, "why.html", page{Title: "Why · Gramlane", Active: "why"})
	})
	mux.HandleFunc("/234", s.framingPage)
	mux.HandleFunc("/framing", s.framingPage)
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
		writeJSON(w, 200, map[string]any{
			"ok": true, "dapp": "gramlane", "layer": "kaspa-l1",
			"unit": "gram", "l2": false, "stablecoin": false,
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
	s.render(w, "home.html", page{Title: "Gramlane — L1 work desk", Active: "home", Jobs: jobs.Catalog})
}

func (s *Server) desk(w http.ResponseWriter, r *http.Request) {
	s.render(w, "desk.html", page{Title: "Desk · Gramlane", Active: "desk", Jobs: jobs.Catalog})
}

func (s *Server) job(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/job/")
	j, ok := jobs.Get(id)
	if !ok {
		http.NotFound(w, r)
		return
	}
	q, err := jobs.QuoteJob(j)
	p := page{Title: j.Name + " · Gramlane", Active: "desk", Job: &j, Query: r.URL.Query().Get("q")}
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

func (s *Server) runPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		http.Redirect(w, r, "/desk", http.StatusSeeOther)
		return
	}
	id := r.FormValue("job")
	q := r.FormValue("q")
	j, ok := jobs.Get(id)
	if !ok {
		s.render(w, "job.html", page{Title: "Run", Active: "desk", Error: "unknown job"})
		return
	}
	paid := strings.TrimSpace(r.FormValue("payment"))
	payer := strings.TrimSpace(r.FormValue("payer"))
	wallet := strings.TrimSpace(r.FormValue("wallet"))
	if paid == "" && payer != "" {
		paid = wallet + ":" + payer
	}
	if paid == "" {
		qq, _ := jobs.QuoteJob(j)
		s.render(w, "job.html", page{
			Title: j.Name, Active: "desk", Job: &j, Query: q, Quote: &qq,
			Error: "Payment required: connect a wallet, or paste a receipt id.",
		})
		return
	}
	rec, err := jobs.RunAs(j, q, paid, payer, wallet)
	p := page{Title: "Receipt · Gramlane", Active: "desk", Job: &j, Query: q, Run: &rec}
	if err != nil {
		p.Error = err.Error()
	}
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
	writeJSON(w, 200, map[string]any{"ok": true, "job": j, "quote": q})
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
	if paid == "" {
		qq, _ := jobs.QuoteJob(j)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusPaymentRequired)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": "Payment Required",
			"note":  "Burn WorkCredit grams (kaspa-work-credit) or pay KAS fallback. Not USD. No L2. Header X-Work-Credit or X-Kaspa-Payment is accepted at HTTP layer only.",
			"job":   j,
			"quote": qq,
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
	writeJSON(w, 200, map[string]any{"ok": true, "receipt": rec})
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

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
