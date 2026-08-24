package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"portfolio-os/internal/models"
	"portfolio-os/internal/services"
	"strconv"
	"strings"
)

type AdminHandler struct {
	Renderer interface {
		Render(http.ResponseWriter, string, any) error
	}
	PortfolioService *services.PortfolioService
}

func NewAdminHandler(
	renderer interface {
		Render(http.ResponseWriter, string, any) error
	},
	portfolioService *services.PortfolioService,
) *AdminHandler {
	return &AdminHandler{
		Renderer:         renderer,
		PortfolioService: portfolioService,
	}
}

func (h *AdminHandler) Login(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodGet {
		if err := h.Renderer.Render(w, "admin-login", nil); err != nil {
			http.Error(
				w,
				"failed to render admin login",
				http.StatusInternalServerError,
			)
		}
		return
	}

	password := r.FormValue("password")

	if password != os.Getenv("ADMIN_PASSWORD") {
		http.Error(
			w,
			"invalid password",
			http.StatusUnauthorized,
		)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "admin_session",
		Value:    os.Getenv("ADMIN_SESSION"),
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
	})

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (h *AdminHandler) Logout(w http.ResponseWriter, r *http.Request) {

	http.SetCookie(w, &http.Cookie{
		Name:     "admin_session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
	})

	http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
}

func (h *AdminHandler) UploadResume(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodPost {
		http.Error(
			w,
			"method not allowed",
			http.StatusMethodNotAllowed,
		)
		return
	}

	if err := saveResume(r); err != nil {
		http.Error(
			w,
			err.Error(),
			http.StatusBadRequest,
		)
		return
	}

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (h *AdminHandler) DownloadResume(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodGet {
		http.Error(
			w,
			"method not allowed",
			http.StatusMethodNotAllowed,
		)
		return
	}

	const resumePath = "uploads/resume/resume.pdf"

	if _, err := os.Stat(resumePath); err != nil {
		if os.IsNotExist(err) {
			http.Error(
				w,
				"resume not found",
				http.StatusNotFound,
			)
			return
		}

		http.Error(
			w,
			"failed to access resume",
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set(
		"Content-Disposition",
		`attachment; filename="resume.pdf"`,
	)

	w.Header().Set(
		"Content-Type",
		"application/pdf",
	)

	http.ServeFile(w, r, resumePath)
}

func (h *AdminHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(
			w,
			"method not allowed",
			http.StatusMethodNotAllowed,
		)
		return
	}

	profile := models.Profile{
		Name:         r.FormValue("name"),
		Title:        r.FormValue("title"),
		Bio:          r.FormValue("bio"),
		Location:     r.FormValue("location"),
		ProfileImage: r.FormValue("profile_image"),
		Availability: r.FormValue("availability"),
	}

	if err := h.PortfolioService.UpdateProfile(profile); err != nil {
		http.Error(
			w,
			"failed to update profile",
			http.StatusInternalServerError,
		)
		return
	}

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (h *AdminHandler) AddCertificate(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(
			w,
			"method not allowed",
			http.StatusMethodNotAllowed,
		)
		return
	}

	imageURL, err := saveCertificateImage(r)
	if err != nil {
		http.Error(
			w,
			"failed to upload certificate image: "+err.Error(),
			http.StatusBadRequest,
		)
		return
	}

	certificate := models.Certificate{
		Title:           r.FormValue("title"),
		Issuer:          r.FormValue("issuer"),
		IssueDate:       r.FormValue("issue_date"),
		CredentialID:    r.FormValue("credential_id"),
		Description:     r.FormValue("description"),
		Image:           imageURL,
		VerificationURL: r.FormValue("verification_url"),
	}

	if certificate.Title == "" || certificate.Issuer == "" {
		http.Error(
			w,
			"title and issuer are required",
			http.StatusBadRequest,
		)
		return
	}

	if err := h.PortfolioService.AddCertificate(certificate); err != nil {
		http.Error(
			w,
			"failed to add certificate",
			http.StatusInternalServerError,
		)
		return
	}

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (h *AdminHandler) UpdateCertificate(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodPost {
		http.Error(
			w,
			"method not allowed",
			http.StatusMethodNotAllowed,
		)
		return
	}

	indexStr := r.PathValue("index")

	var index int

	if _, err := fmt.Sscanf(indexStr, "%d", &index); err != nil {
		http.Error(
			w,
			"invalid certificate index",
			http.StatusBadRequest,
		)
		return
	}

	currentPortfolio := h.PortfolioService.GetPortfolio()

	if index < 0 || index >= len(currentPortfolio.Certificates) {
		http.Error(
			w,
			"certificate not found",
			http.StatusNotFound,
		)
		return
	}

	currentCertificate := currentPortfolio.Certificates[index]

	imageURL, err := saveCertificateImage(r)
	if err != nil {
		http.Error(
			w,
			"failed to upload certificate image: "+err.Error(),
			http.StatusBadRequest,
		)
		return
	}

	// Keep the existing image if no new image was uploaded.
	if imageURL == "" {
		imageURL = currentCertificate.Image
	}

	certificate := models.Certificate{
		Title:           r.FormValue("title"),
		Issuer:          r.FormValue("issuer"),
		IssueDate:       r.FormValue("issue_date"),
		CredentialID:    r.FormValue("credential_id"),
		Description:     r.FormValue("description"),
		Image:           imageURL,
		VerificationURL: r.FormValue("verification_url"),
	}

	if certificate.Title == "" || certificate.Issuer == "" {
		http.Error(
			w,
			"title and issuer are required",
			http.StatusBadRequest,
		)
		return
	}

	if err := h.PortfolioService.UpdateCertificate(
		index,
		certificate,
	); err != nil {
		http.Error(
			w,
			"failed to update certificate",
			http.StatusInternalServerError,
		)
		return
	}

	http.Redirect(
		w,
		r,
		"/admin",
		http.StatusSeeOther,
	)
}

func (h *AdminHandler) DeleteCertificate(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodPost {
		http.Error(
			w,
			"method not allowed",
			http.StatusMethodNotAllowed,
		)
		return
	}

	indexStr := r.PathValue("index")

	var index int

	if _, err := fmt.Sscanf(indexStr, "%d", &index); err != nil {
		http.Error(
			w,
			"invalid certificate index",
			http.StatusBadRequest,
		)
		return
	}

	if err := h.PortfolioService.DeleteCertificate(index); err != nil {
		http.Error(
			w,
			"failed to delete certificate",
			http.StatusInternalServerError,
		)
		return
	}

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (h *AdminHandler) AddSkill(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(
			w,
			"method not allowed",
			http.StatusMethodNotAllowed,
		)
		return
	}

	skill, ok := parseSkillForm(w, r)
	if !ok {
		return
	}

	if err := h.PortfolioService.AddSkill(skill); err != nil {
		http.Error(
			w,
			"failed to add skill",
			http.StatusInternalServerError,
		)
		return
	}

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (h *AdminHandler) UpdateSkill(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodPost {
		http.Error(
			w,
			"method not allowed",
			http.StatusMethodNotAllowed,
		)
		return
	}

	indexStr := r.PathValue("index")

	var index int

	if _, err := fmt.Sscanf(indexStr, "%d", &index); err != nil {
		http.Error(
			w,
			"invalid skill index",
			http.StatusBadRequest,
		)
		return
	}

	skill, ok := parseSkillForm(w, r)
	if !ok {
		return
	}

	if err := h.PortfolioService.UpdateSkill(index, skill); err != nil {
		http.Error(
			w,
			"failed to update skill",
			http.StatusInternalServerError,
		)
		return
	}

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (h *AdminHandler) DeleteSkill(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodPost {
		http.Error(
			w,
			"method not allowed",
			http.StatusMethodNotAllowed,
		)
		return
	}

	indexStr := r.PathValue("index")

	var index int

	if _, err := fmt.Sscanf(indexStr, "%d", &index); err != nil {
		http.Error(
			w,
			"invalid skill index",
			http.StatusBadRequest,
		)
		return
	}

	if err := h.PortfolioService.DeleteSkill(index); err != nil {
		http.Error(
			w,
			"failed to delete skill",
			http.StatusInternalServerError,
		)
		return
	}

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (h *AdminHandler) AddStatistic(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(
			w,
			"method not allowed",
			http.StatusMethodNotAllowed,
		)
		return
	}

	statistic, ok := parseStatisticForm(w, r)
	if !ok {
		return
	}

	if err := h.PortfolioService.AddStatistic(statistic); err != nil {
		http.Error(
			w,
			"failed to add statistic",
			http.StatusInternalServerError,
		)
		return
	}

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (h *AdminHandler) UpdateStatistic(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodPost {
		http.Error(
			w,
			"method not allowed",
			http.StatusMethodNotAllowed,
		)
		return
	}

	indexStr := r.PathValue("index")

	var index int

	if _, err := fmt.Sscanf(indexStr, "%d", &index); err != nil {
		http.Error(
			w,
			"invalid statistic index",
			http.StatusBadRequest,
		)
		return
	}

	statistic, ok := parseStatisticForm(w, r)
	if !ok {
		return
	}

	if err := h.PortfolioService.UpdateStatistic(index, statistic); err != nil {
		http.Error(
			w,
			"failed to update statistic",
			http.StatusInternalServerError,
		)
		return
	}

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (h *AdminHandler) DeleteStatistic(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodPost {
		http.Error(
			w,
			"method not allowed",
			http.StatusMethodNotAllowed,
		)
		return
	}

	indexStr := r.PathValue("index")

	var index int

	if _, err := fmt.Sscanf(indexStr, "%d", &index); err != nil {
		http.Error(
			w,
			"invalid statistic index",
			http.StatusBadRequest,
		)
		return
	}

	if err := h.PortfolioService.DeleteStatistic(index); err != nil {
		http.Error(
			w,
			"failed to delete statistic",
			http.StatusInternalServerError,
		)
		return
	}

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (h *AdminHandler) AddService(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(
			w,
			"method not allowed",
			http.StatusMethodNotAllowed,
		)
		return
	}

	service := models.ServiceOffering{
		Title:       r.FormValue("title"),
		Description: r.FormValue("description"),
		Icon:        r.FormValue("icon"),
	}

	if service.Title == "" || service.Description == "" {
		http.Error(
			w,
			"title and description are required",
			http.StatusBadRequest,
		)
		return
	}

	if err := h.PortfolioService.AddService(service); err != nil {
		http.Error(
			w,
			"failed to add service",
			http.StatusInternalServerError,
		)
		return
	}

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (h *AdminHandler) UpdateService(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodPost {
		http.Error(
			w,
			"method not allowed",
			http.StatusMethodNotAllowed,
		)
		return
	}

	indexStr := r.PathValue("index")

	var index int

	if _, err := fmt.Sscanf(indexStr, "%d", &index); err != nil {
		http.Error(
			w,
			"invalid service index",
			http.StatusBadRequest,
		)
		return
	}

	service := models.ServiceOffering{
		Title:       r.FormValue("title"),
		Description: r.FormValue("description"),
		Icon:        r.FormValue("icon"),
	}

	if service.Title == "" || service.Description == "" {
		http.Error(
			w,
			"title and description are required",
			http.StatusBadRequest,
		)
		return
	}

	if err := h.PortfolioService.UpdateService(index, service); err != nil {
		http.Error(
			w,
			"failed to update service",
			http.StatusInternalServerError,
		)
		return
	}

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (h *AdminHandler) DeleteService(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodPost {
		http.Error(
			w,
			"method not allowed",
			http.StatusMethodNotAllowed,
		)
		return
	}

	indexStr := r.PathValue("index")

	var index int

	if _, err := fmt.Sscanf(indexStr, "%d", &index); err != nil {
		http.Error(
			w,
			"invalid service index",
			http.StatusBadRequest,
		)
		return
	}

	if err := h.PortfolioService.DeleteService(index); err != nil {
		http.Error(
			w,
			"failed to delete service",
			http.StatusInternalServerError,
		)
		return
	}

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (h *AdminHandler) AddEducation(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(
			w,
			"method not allowed",
			http.StatusMethodNotAllowed,
		)
		return
	}

	entry, ok := parseEducationForm(w, r)
	if !ok {
		return
	}

	if err := h.PortfolioService.AddEducation(entry); err != nil {
		http.Error(
			w,
			"failed to add education entry",
			http.StatusInternalServerError,
		)
		return
	}

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (h *AdminHandler) UpdateEducation(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodPost {
		http.Error(
			w,
			"method not allowed",
			http.StatusMethodNotAllowed,
		)
		return
	}

	indexStr := r.PathValue("index")

	var index int

	if _, err := fmt.Sscanf(indexStr, "%d", &index); err != nil {
		http.Error(
			w,
			"invalid education index",
			http.StatusBadRequest,
		)
		return
	}

	entry, ok := parseEducationForm(w, r)
	if !ok {
		return
	}

	if err := h.PortfolioService.UpdateEducation(index, entry); err != nil {
		http.Error(
			w,
			"failed to update education entry",
			http.StatusInternalServerError,
		)
		return
	}

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (h *AdminHandler) DeleteEducation(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodPost {
		http.Error(
			w,
			"method not allowed",
			http.StatusMethodNotAllowed,
		)
		return
	}

	indexStr := r.PathValue("index")

	var index int

	if _, err := fmt.Sscanf(indexStr, "%d", &index); err != nil {
		http.Error(
			w,
			"invalid education index",
			http.StatusBadRequest,
		)
		return
	}

	if err := h.PortfolioService.DeleteEducation(index); err != nil {
		http.Error(
			w,
			"failed to delete education entry",
			http.StatusInternalServerError,
		)
		return
	}

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (h *AdminHandler) AddProject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(
			w,
			"method not allowed",
			http.StatusMethodNotAllowed,
		)
		return
	}

	project, ok := parseProjectForm(w, r)
	if !ok {
		return
	}

	imageURL, err := saveProjectImage(r)
	if err != nil {
		http.Error(
			w,
			"failed to upload project image: "+err.Error(),
			http.StatusBadRequest,
		)
		return
	}

	project.Image = imageURL

	if err := h.PortfolioService.AddProject(project); err != nil {
		http.Error(
			w,
			"failed to add project",
			http.StatusInternalServerError,
		)
		return
	}

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (h *AdminHandler) UpdateProject(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodPost {
		http.Error(
			w,
			"method not allowed",
			http.StatusMethodNotAllowed,
		)
		return
	}

	indexStr := r.PathValue("index")

	var index int

	if _, err := fmt.Sscanf(indexStr, "%d", &index); err != nil {
		http.Error(
			w,
			"invalid project index",
			http.StatusBadRequest,
		)
		return
	}

	currentPortfolio := h.PortfolioService.GetPortfolio()

	if index < 0 || index >= len(currentPortfolio.Projects) {
		http.Error(
			w,
			"invalid project index",
			http.StatusBadRequest,
		)
		return
	}

	project, ok := parseProjectForm(w, r)
	if !ok {
		return
	}

	project.Image = currentPortfolio.Projects[index].Image

	imageURL, err := saveProjectImage(r)
	if err != nil {
		http.Error(
			w,
			"failed to upload project image: "+err.Error(),
			http.StatusBadRequest,
		)
		return
	}

	if imageURL != "" {
		project.Image = imageURL
	}

	if err := h.PortfolioService.UpdateProject(index, project); err != nil {
		http.Error(
			w,
			"failed to update project",
			http.StatusInternalServerError,
		)
		return
	}

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (h *AdminHandler) DeleteProject(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodPost {
		http.Error(
			w,
			"method not allowed",
			http.StatusMethodNotAllowed,
		)
		return
	}

	indexStr := r.PathValue("index")

	var index int

	if _, err := fmt.Sscanf(indexStr, "%d", &index); err != nil {
		http.Error(
			w,
			"invalid project index",
			http.StatusBadRequest,
		)
		return
	}

	if err := h.PortfolioService.DeleteProject(index); err != nil {
		http.Error(
			w,
			"failed to delete project",
			http.StatusInternalServerError,
		)
		return
	}

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (h *AdminHandler) UpdateContact(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(
			w,
			"method not allowed",
			http.StatusMethodNotAllowed,
		)
		return
	}

	contact := models.ContactInfo{
		Email:    r.FormValue("email"),
		Phone:    r.FormValue("phone"),
		Location: r.FormValue("location"),
	}

	if err := h.PortfolioService.UpdateContact(contact); err != nil {
		http.Error(
			w,
			"failed to update contact",
			http.StatusInternalServerError,
		)
		return
	}

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (h *AdminHandler) AddSocialLink(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(
			w,
			"method not allowed",
			http.StatusMethodNotAllowed,
		)
		return
	}

	link, ok := parseSocialLinkForm(w, r)
	if !ok {
		return
	}

	if err := h.PortfolioService.AddSocialLink(link); err != nil {
		http.Error(
			w,
			"failed to add social link",
			http.StatusInternalServerError,
		)
		return
	}

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (h *AdminHandler) UpdateSocialLink(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodPost {
		http.Error(
			w,
			"method not allowed",
			http.StatusMethodNotAllowed,
		)
		return
	}

	indexStr := r.PathValue("index")

	var index int

	if _, err := fmt.Sscanf(indexStr, "%d", &index); err != nil {
		http.Error(
			w,
			"invalid social link index",
			http.StatusBadRequest,
		)
		return
	}

	link, ok := parseSocialLinkForm(w, r)
	if !ok {
		return
	}

	if err := h.PortfolioService.UpdateSocialLink(index, link); err != nil {
		http.Error(
			w,
			"failed to update social link",
			http.StatusInternalServerError,
		)
		return
	}

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (h *AdminHandler) DeleteSocialLink(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodPost {
		http.Error(
			w,
			"method not allowed",
			http.StatusMethodNotAllowed,
		)
		return
	}

	indexStr := r.PathValue("index")

	var index int

	if _, err := fmt.Sscanf(indexStr, "%d", &index); err != nil {
		http.Error(
			w,
			"invalid social link index",
			http.StatusBadRequest,
		)
		return
	}

	if err := h.PortfolioService.DeleteSocialLink(index); err != nil {
		http.Error(
			w,
			"failed to delete social link",
			http.StatusInternalServerError,
		)
		return
	}

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

// parseSkillForm builds a Skill from the request form, validating that a name is
// present and that proficiency is a number between 0 and 100. On failure it writes
// a 400 response and returns ok == false.
func parseSkillForm(
	w http.ResponseWriter,
	r *http.Request,
) (models.Skill, bool) {
	name := r.FormValue("name")

	if name == "" {
		http.Error(
			w,
			"skill name is required",
			http.StatusBadRequest,
		)
		return models.Skill{}, false
	}

	proficiency, err := strconv.Atoi(r.FormValue("proficiency"))
	if err != nil || proficiency < 0 || proficiency > 100 {
		http.Error(
			w,
			"proficiency must be a number between 0 and 100",
			http.StatusBadRequest,
		)
		return models.Skill{}, false
	}

	return models.Skill{
		Name:        name,
		Category:    r.FormValue("category"),
		Proficiency: proficiency,
	}, true
}

// parseEducationForm builds an EducationEntry from the request form. Institution and
// Degree are required; Field and Description are optional. StartYear and EndYear are
// optional (blank is stored as 0, representing ongoing studies) but, when provided,
// must be valid years. On failure it writes a 400 response and returns ok == false.
func parseEducationForm(
	w http.ResponseWriter,
	r *http.Request,
) (models.EducationEntry, bool) {
	institution := r.FormValue("institution")
	degree := r.FormValue("degree")

	if institution == "" || degree == "" {
		http.Error(
			w,
			"institution and degree are required",
			http.StatusBadRequest,
		)
		return models.EducationEntry{}, false
	}

	startYear, ok := parseOptionalYear(w, r.FormValue("start_year"), "start year")
	if !ok {
		return models.EducationEntry{}, false
	}

	endYear, ok := parseOptionalYear(w, r.FormValue("end_year"), "end year")
	if !ok {
		return models.EducationEntry{}, false
	}

	if startYear != 0 && endYear != 0 && endYear < startYear {
		http.Error(
			w,
			"end year must not be before start year",
			http.StatusBadRequest,
		)
		return models.EducationEntry{}, false
	}

	return models.EducationEntry{
		Institution: institution,
		Degree:      degree,
		Field:       r.FormValue("field"),
		StartYear:   startYear,
		EndYear:     endYear,
		Description: r.FormValue("description"),
	}, true
}

// parseOptionalYear parses a year form value. A blank value is allowed and returns 0
// (used to represent an unset or ongoing year). A non-empty value must be an integer
// within a sensible range.
func parseOptionalYear(
	w http.ResponseWriter,
	value string,
	label string,
) (int, bool) {
	if value == "" {
		return 0, true
	}

	year, err := strconv.Atoi(value)
	if err != nil || year < 1950 || year > 2100 {
		http.Error(
			w,
			label+" must be a year between 1950 and 2100",
			http.StatusBadRequest,
		)
		return 0, false
	}

	return year, true
}

// parseProjectForm builds a Project from the request form. Name and Description are
// required; Image, URL, and GitHub are optional. Technologies is a comma-separated
// list, trimmed with empty entries dropped. On failure it writes a 400 response and
// returns ok == false.
func parseProjectForm(
	w http.ResponseWriter,
	r *http.Request,
) (models.Project, bool) {
	name := r.FormValue("name")
	description := r.FormValue("description")

	if name == "" || description == "" {
		http.Error(
			w,
			"name and description are required",
			http.StatusBadRequest,
		)
		return models.Project{}, false
	}

	var technologies []string

	for _, technology := range strings.Split(r.FormValue("technologies"), ",") {
		technology = strings.TrimSpace(technology)
		if technology != "" {
			technologies = append(technologies, technology)
		}
	}

	return models.Project{
		Name:         name,
		Description:  description,
		URL:          r.FormValue("url"),
		GitHub:       r.FormValue("github"),
		Technologies: technologies,
	}, true
}

// parseSocialLinkForm builds a SocialLink from the request form. Platform and URL are
// required; Username is optional. On failure it writes a 400 response and returns
// ok == false.
func parseSocialLinkForm(
	w http.ResponseWriter,
	r *http.Request,
) (models.SocialLink, bool) {
	platform := r.FormValue("platform")
	url := r.FormValue("url")

	if platform == "" || url == "" {
		http.Error(
			w,
			"platform and URL are required",
			http.StatusBadRequest,
		)
		return models.SocialLink{}, false
	}

	return models.SocialLink{
		Platform: platform,
		URL:      url,
		Username: r.FormValue("username"),
	}, true
}

func parseStatisticForm(
	w http.ResponseWriter,
	r *http.Request,
) (models.Statistic, bool) {
	label := r.FormValue("label")
	value := r.FormValue("value")

	if label == "" || value == "" {
		http.Error(
			w,
			"statistic label and value are required",
			http.StatusBadRequest,
		)
		return models.Statistic{}, false
	}

	return models.Statistic{
		Label: label,
		Value: value,
	}, true
}

func saveCertificateImage(r *http.Request) (string, error) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		return "", fmt.Errorf("failed to parse upload: %w", err)
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		if err == http.ErrMissingFile {
			return "", nil
		}

		return "", fmt.Errorf("failed to read image: %w", err)
	}
	defer file.Close()

	if header.Size > 10<<20 {
		return "", fmt.Errorf("image is too large; maximum size is 10MB")
	}

	buffer := make([]byte, 512)

	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("failed to inspect image: %w", err)
	}

	contentType := http.DetectContentType(buffer[:n])

	extensions := map[string]string{
		"image/jpeg": ".jpg",
		"image/png":  ".png",
		"image/webp": ".webp",
	}

	extension, ok := extensions[contentType]
	if !ok {
		return "", fmt.Errorf("unsupported image type: %s", contentType)
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", fmt.Errorf("failed to reset image: %w", err)
	}

	if err := os.MkdirAll("uploads/certificates", 0755); err != nil {
		return "", fmt.Errorf("failed to create upload directory: %w", err)
	}

	imageID, err := services.GenerateVisitID()
	if err != nil {
		return "", fmt.Errorf("failed to generate image ID: %w", err)
	}

	filename := imageID + extension

	path := filepath.Join(
		"uploads",
		"certificates",
		filename,
	)

	output, err := os.Create(path)
	if err != nil {
		return "", fmt.Errorf("failed to create image: %w", err)
	}
	defer output.Close()

	if _, err := io.Copy(output, file); err != nil {
		return "", fmt.Errorf("failed to save image: %w", err)
	}

	return "/uploads/certificates/" + filename, nil
}

func saveProjectImage(r *http.Request) (string, error) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		return "", fmt.Errorf("failed to parse upload: %w", err)
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		if err == http.ErrMissingFile {
			return "", nil
		}

		return "", fmt.Errorf("failed to read image: %w", err)
	}
	defer file.Close()

	if header.Size > 10<<20 {
		return "", fmt.Errorf("image is too large; maximum size is 10MB")
	}

	buffer := make([]byte, 512)

	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("failed to inspect image: %w", err)
	}

	contentType := http.DetectContentType(buffer[:n])

	extensions := map[string]string{
		"image/jpeg": ".jpg",
		"image/png":  ".png",
		"image/webp": ".webp",
	}

	extension, ok := extensions[contentType]
	if !ok {
		return "", fmt.Errorf("unsupported image type: %s", contentType)
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", fmt.Errorf("failed to reset image: %w", err)
	}

	if err := os.MkdirAll("uploads/projects", 0755); err != nil {
		return "", fmt.Errorf("failed to create upload directory: %w", err)
	}

	imageID, err := services.GenerateVisitID()
	if err != nil {
		return "", fmt.Errorf("failed to generate image ID: %w", err)
	}

	filename := imageID + extension

	path := filepath.Join(
		"uploads",
		"projects",
		filename,
	)

	output, err := os.Create(path)
	if err != nil {
		return "", fmt.Errorf("failed to create image: %w", err)
	}
	defer output.Close()

	if _, err := io.Copy(output, file); err != nil {
		return "", fmt.Errorf("failed to save image: %w", err)
	}

	return "/uploads/projects/" + filename, nil
}

func saveResume(r *http.Request) error {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		return fmt.Errorf("failed to parse upload: %w", err)
	}

	file, header, err := r.FormFile("resume")
	if err != nil {
		if err == http.ErrMissingFile {
			return fmt.Errorf("resume file is required")
		}

		return fmt.Errorf("failed to read resume: %w", err)
	}
	defer file.Close()

	if header.Size > 10<<20 {
		return fmt.Errorf("resume is too large; maximum size is 10MB")
	}

	buffer := make([]byte, 512)

	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return fmt.Errorf("failed to inspect resume: %w", err)
	}

	contentType := http.DetectContentType(buffer[:n])

	if contentType != "application/pdf" {
		return fmt.Errorf("unsupported resume type: %s; PDF is required", contentType)
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("failed to reset resume: %w", err)
	}

	if err := os.MkdirAll("uploads/resume", 0755); err != nil {
		return fmt.Errorf("failed to create resume directory: %w", err)
	}

	path := filepath.Join(
		"uploads",
		"resume",
		"resume.pdf",
	)

	output, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create resume: %w", err)
	}
	defer output.Close()

	if _, err := io.Copy(output, file); err != nil {
		return fmt.Errorf("failed to save resume: %w", err)
	}

	return nil
}
