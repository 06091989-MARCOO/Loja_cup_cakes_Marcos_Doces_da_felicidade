package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
)

type Cupcake struct {
	ID        int     `json:"id"`
	Nome      string  `json:"nome"`
	Sabor     string  `json:"sabor"`
	Preco     float64 `json:"preco"`
	Descricao string  `json:"descricao"`
	Imagem    string  `json:"imagem"`
}

type ItemCarrinho struct {
	Cupcake    Cupcake `json:"cupcake"`
	Quantidade int     `json:"quantidade"`
}

var (
	catalogo = []Cupcake{
		{ID: 1, Nome: "Red Velvet", Sabor: "Baunilha com cacau", Preco: 8.50, Descricao: "Clássico red velvet com cobertura de cream cheese.", Imagem: "https://unsplash.com"},
		{ID: 2, Nome: "ChocoDelícia", Sabor: "Chocolate", Preco: 9.00, Descricao: "Massa de chocolate belga com recheio de brigadeiro.", Imagem: "https://unsplash.com"},
		{ID: 3, Nome: "Ninho com Nutella", Sabor: "Leite Ninho", Preco: 10.50, Descricao: "Massa de leite em pó recheada com Nutella pura.", Imagem: "https://unsplash.com"},
	}
	carrinho = []ItemCarrinho{}
	mu       sync.Mutex
)

func main() {
	// Rota do Site Principal
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, "index.html")
	})

	// NOVA ROTA: Validação de Login Simples
	http.HandleFunc("/api/login", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != "POST" {
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
			return
		}

		var credenciais struct {
			Usuario string `json:"usuario"`
			Senha   string `json:"senha"`
		}

		if err := json.NewDecoder(r.Body).Decode(&credenciais); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"erro": "Dados inválidos"})
			return
		}

		// Usuário e senha padrão configurados aqui na memória
		if credenciais.Usuario == "admin" && credenciais.Senha == "1234" {
			resposta := map[string]string{
				"nome":  "Administrador",
				"token": "token-ficticio-sucesso-abc123xyz",
				"msg":   "Logado com sucesso!",
			}
			json.NewEncoder(w).Encode(resposta)
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"erro": "Usuário ou senha incorretos!"})
		}
	})

	http.HandleFunc("/api/cupcakes", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(catalogo)
	})

	http.HandleFunc("/api/carrinho", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == "GET" {
			mu.Lock()
			json.NewEncoder(w).Encode(carrinho)
			mu.Unlock()
			return
		}
		if r.Method == "POST" {
			var req struct {
				CupcakeID  int `json:"cupcake_id"`
				Quantidade int `json:"quantidade"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "Invalido", http.StatusBadRequest)
				return
			}
			var cFound *Cupcake
			for _, c := range catalogo {
				if c.ID == req.CupcakeID {
					cFound = &c
					break
				}
			}
			if cFound == nil {
				http.Error(w, "Nao encontrado", http.StatusNotFound)
				return
			}
			mu.Lock()
			added := false
			for i, item := range carrinho {
				if item.Cupcake.ID == cFound.ID {
					carrinho[i].Quantidade += req.Quantidade
					added = true
					break
				}
			}
			if !added {
				carrinho = append(carrinho, ItemCarrinho{Cupcake: *cFound, Quantidade: req.Quantidade})
			}
			mu.Unlock()
			w.Write([]byte(`{"mensagem": "Adicionado!"}`))
		}
	})

	http.HandleFunc("/api/checkout", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		mu.Lock()
		defer mu.Unlock()
		if len(carrinho) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"erro": "Carrinho vazio"}`))
			return
		}
		var total float64
		for _, item := range carrinho {
			total += item.Cupcake.Preco * float64(item.Quantidade)
		}
		carrinho = []ItemCarrinho{}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"mensagem": "Pedido realizado com sucesso! 🧁",
			"total":    total,
		})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Printf("🧁 Confeitaria '%s' ATIVA na porta %s!\n", "Marcos doces da felicidade", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

