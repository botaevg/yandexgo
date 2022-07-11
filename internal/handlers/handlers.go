package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/botaevg/yandexgo/internal/config"
	"github.com/botaevg/yandexgo/internal/domain"
	"github.com/botaevg/yandexgo/internal/repositories"
	"github.com/botaevg/yandexgo/internal/shorten"
	"github.com/go-chi/chi/v5"
	"io"
	"log"
	"net/http"
	"net/url"
)

type handler struct {
	config         config.Config
	storage        repositories.Storage
	asyncExecution chan DeleteURL
}

func New(cfg config.Config, storage repositories.Storage, asyncExecution chan DeleteURL) *handler {
	return &handler{
		config:         cfg,
		storage:        storage,
		asyncExecution: asyncExecution,
	}
}

type URL struct {
	FullURL  string `json:"url"`
	ShortURL string `json:"result"`
}

type DeleteURL struct {
	shorts []string
	idUser string
}

func (h *handler) APIDelete(w http.ResponseWriter, r *http.Request) {
	// update urls set deleted = 100 where shortURL = []shorts
	//idUser := cookies.VerificationCookie(h.storage, r, &w)
	idUser := r.Context().Value("idUser").(string)
	log.Print(idUser)
	b, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var shorts []string
	if err := json.Unmarshal(b, &shorts); err != nil {
		log.Print("Unmarshal error")
		log.Print(err)
		http.Error(w, errors.New("BadRequest").Error(), http.StatusBadRequest)
		return
	}

	h.asyncExecution <- DeleteURL{
		shorts: shorts,
		idUser: idUser,
	}

	log.Print(shorts)
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("oke"))
}

func (h *handler) DeleteAsync(d DeleteURL) {
	err := h.storage.UpdateFlagDelete(d.shorts, d.idUser)
	if err != nil {
		log.Print("ошибка обновления удаления")

	}
}

func (h *handler) APIShortBatch(w http.ResponseWriter, r *http.Request) {
	//idUser := cookies.VerificationCookie(h.storage, r, &w)
	idUser := r.Context().Value("idUser").(string)
	log.Print(idUser)

	b, err := io.ReadAll(r.Body) //reader
	// обрабатываем ошибку
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var originBatch []domain.APIOriginBatch
	if err := json.Unmarshal(b, &originBatch); err != nil {
		log.Print("Unmarshal error")
		log.Print(err)
		http.Error(w, errors.New("BadRequest").Error(), http.StatusBadRequest)
		return
	}

	var URLForAddStorage []domain.URLForAddStorage
	var shortBatch []domain.APIShortBatch

	for _, v := range originBatch {
		shortNew := shorten.ShortURL()
		URLForAddStorage = append(URLForAddStorage, domain.URLForAddStorage{
			FullURL:  v.Origin,
			IDUser:   idUser,
			ShortURL: shortNew,
		})
		shortBatch = append(shortBatch, domain.APIShortBatch{
			ID:       v.ID,
			ShortURL: h.config.BaseURL + shortNew,
		})
	}

	err = h.storage.AddShortBatch(URLForAddStorage)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	/*for _, v := range URLForAddStorage {
		shortBatch = append(shortBatch, domain.APIShortBatch{
			ID:       v.CorrelationID,
			ShortURL: h.config.BaseURL + v.ShortURL,
		})
	} // можно сразу формировать выше*/

	b, err = json.Marshal(shortBatch)
	if err != nil {
		log.Print("Marshal error")
		log.Print(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(b)
}

func (h *handler) CheckPing(w http.ResponseWriter, r *http.Request) {

	if err := h.storage.Ping(context.Background()); err != nil {
		log.Print("ping error")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("oki"))
}

func (h *handler) GetAllShortURL(w http.ResponseWriter, r *http.Request) {
	//idUser := cookies.VerificationCookie(h.storage, r, &w)
	idUser := r.Context().Value("idUser").(string)
	log.Print(idUser)

	URLForGetAll, err := h.storage.GetAllShort(idUser)

	var allShortURL []domain.URLpair
	for _, v := range URLForGetAll {
		allShortURL = append(allShortURL, domain.URLpair{
			FullURL:  v.FullURL,
			ShortURL: h.config.BaseURL + v.ShortURL,
		})
	}

	if err != nil {
		w.WriteHeader(http.StatusNoContent)
		w.Write([]byte("Не найдено"))
		return
	}
	b, err := json.Marshal(allShortURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	log.Print(b)
	log.Print(string(b))
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

func (h *handler) GetHandler(w http.ResponseWriter, r *http.Request) {
	//idUser := cookies.VerificationCookie(h.storage, r, &w)
	idUser := r.Context().Value("idUser").(string)
	log.Print(idUser)

	id := chi.URLParam(r, "id")

	u, err := h.storage.GetFullURL(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if u.Deleted {
		w.WriteHeader(http.StatusGone)
		w.Write([]byte("Gone"))
		return
	}
	if u.FullURL == "" {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Не найдено"))
		return
	}

	w.Header().Set("Location", u.FullURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
	w.Write([]byte(u.FullURL))

}

func (h *handler) PostHandler(w http.ResponseWriter, r *http.Request) {
	//idUser := cookies.VerificationCookie(h.storage, r, &w)
	idUser := r.Context().Value("idUser").(string)
	log.Print(idUser)

	b, err := io.ReadAll(r.Body) //reader
	// обрабатываем ошибку
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	strURL := string(b)
	_, err = url.ParseRequestURI(strURL)
	if err != nil {
		http.Error(w, errors.New("BadRequest").Error(), http.StatusBadRequest)
		return
	}

	//shortURLs := shorten.ShortURL()
	//err = h.storage.AddShort(strURL, shortURLs, idUser)
	shortURLs, newShort, err := AddOrFindURL(h.storage, strURL, idUser)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	baseURL := h.config.BaseURL // os.LookupEnv("BASE_URL")
	if newShort {
		w.WriteHeader(http.StatusCreated)
	} else {
		w.WriteHeader(http.StatusConflict)
	}
	w.Write([]byte(baseURL + shortURLs))

}

func (h *handler) APIPost(w http.ResponseWriter, r *http.Request) {
	//idUser := cookies.VerificationCookie(h.storage, r, &w)
	idUser := r.Context().Value("idUser").(string)
	log.Print(idUser)

	b, err := io.ReadAll(r.Body) //reader
	// обрабатываем ошибку
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	obj := URL{}
	if err := json.Unmarshal(b, &obj); err != nil {
		http.Error(w, errors.New("BadRequest").Error(), http.StatusBadRequest)
		return
	}
	strURL := obj.FullURL

	_, err = url.ParseRequestURI(strURL)
	if err != nil {
		http.Error(w, errors.New("BadRequest").Error(), http.StatusBadRequest)
		return
	}

	//shortURLs := shorten.ShortURL()
	//err = h.storage.AddShort(strURL, shortURLs, idUser)
	shortURLs, newShort, err := AddOrFindURL(h.storage, strURL, idUser)
	if err != nil {

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	baseURL := h.config.BaseURL // os.LookupEnv("BASE_URL")

	w.Header().Set("content-type", "application/json")
	if newShort {
		w.WriteHeader(http.StatusCreated)
	} else {
		w.WriteHeader(http.StatusConflict)
	}

	w.Write([]byte(`{"result":"` + baseURL + shortURLs + `"}`))
}

func AddOrFindURL(h repositories.Storage, strURL string, idUser string) (string, bool, error) {
	newShort := true
	shortURLs := shorten.ShortURL()
	shortURLsAfter, err := h.AddShort(domain.URLForAddStorage{
		FullURL:  strURL,
		ShortURL: shortURLs,
		IDUser:   idUser,
	})
	if err != nil {
		return "", true, err

	}
	if shortURLs != shortURLsAfter {
		shortURLs = shortURLsAfter
		newShort = false
	}
	return shortURLs, newShort, nil
}
