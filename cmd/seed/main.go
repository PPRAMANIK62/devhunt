// cmd/seed/main.go — populates the database with realistic fixture data.
// Run with: go run ./cmd/seed
// Safe to re-run: existing seed rows are deleted first.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/PPRAMANIK62/devhunt/internal/config"
	"github.com/PPRAMANIK62/devhunt/internal/database"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()
	db, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "db: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := seed(ctx, db); err != nil {
		fmt.Fprintf(os.Stderr, "seed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("seed complete")
}

func hash(password string) string {
	b, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		panic(err)
	}
	return string(b)
}

type company struct {
	id   string
	slug string
}

func insertCompany(ctx context.Context, db *pgxpool.Pool, email, name, slug, description, website string) (company, error) {
	var userID string
	err := db.QueryRow(ctx, `
		INSERT INTO users (email, password_hash, role)
		VALUES ($1, $2, 'company') RETURNING id
	`, email, hash("password123")).Scan(&userID)
	if err != nil {
		return company{}, fmt.Errorf("insert user %s: %w", email, err)
	}

	var id string
	err = db.QueryRow(ctx, `
		INSERT INTO companies (user_id, name, slug, description, website)
		VALUES ($1, $2, $3, $4, $5) RETURNING id
	`, userID, name, slug, description, website).Scan(&id)
	if err != nil {
		return company{}, fmt.Errorf("insert company %s: %w", slug, err)
	}

	return company{id: id, slug: slug}, nil
}

func insertJob(ctx context.Context, db *pgxpool.Pool, companyID, title, description, location string, salaryMin, salaryMax int, status string, tags []string) error {
	var jobID string
	err := db.QueryRow(ctx, `
		INSERT INTO jobs (company_id, title, description, location, salary_min, salary_max, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id
	`, companyID, title, description, location, salaryMin, salaryMax, status).Scan(&jobID)
	if err != nil {
		return fmt.Errorf("insert job %q: %w", title, err)
	}

	for _, tag := range tags {
		var tagID string
		if err := db.QueryRow(ctx, `
			INSERT INTO tags (name) VALUES ($1)
			ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
			RETURNING id
		`, tag).Scan(&tagID); err != nil {
			return fmt.Errorf("upsert tag %q: %w", tag, err)
		}
		if _, err := db.Exec(ctx, `
			INSERT INTO job_tags (job_id, tag_id) VALUES ($1, $2) ON CONFLICT DO NOTHING
		`, jobID, tagID); err != nil {
			return fmt.Errorf("link tag %q: %w", tag, err)
		}
	}

	fmt.Printf("  job: %s [%s]\n", title, status)
	return nil
}

func seed(ctx context.Context, db *pgxpool.Pool) error {
	// ── Wipe previous seed data ──────────────────────────────────────────
	slugs := []string{
		"acme-corp", "devstudio", "finledger", "neural-labs",
		"forge-tools", "pixel-agency", "cloudnine",
	}
	emails := []string{
		"acme@example.com", "devstudio@example.com", "finledger@example.com",
		"neural-labs@example.com", "forge-tools@example.com",
		"pixel-agency@example.com", "cloudnine@example.com",
		"seeker@example.com",
	}

	slugPlaceholders := ""
	emailPlaceholders := ""
	slugArgs := make([]any, len(slugs))
	emailArgs := make([]any, len(emails))
	for i, s := range slugs {
		if i > 0 {
			slugPlaceholders += ", "
		}
		slugPlaceholders += fmt.Sprintf("$%d", i+1)
		slugArgs[i] = s
	}
	for i, e := range emails {
		if i > 0 {
			emailPlaceholders += ", "
		}
		emailPlaceholders += fmt.Sprintf("$%d", i+1)
		emailArgs[i] = e
	}

	if _, err := db.Exec(ctx, fmt.Sprintf(`
		DELETE FROM applications WHERE user_id IN (
			SELECT id FROM users WHERE email IN (%s)
		)`, emailPlaceholders), emailArgs...); err != nil {
		return fmt.Errorf("wipe applications: %w", err)
	}
	if _, err := db.Exec(ctx, fmt.Sprintf(`
		DELETE FROM job_tags WHERE job_id IN (
			SELECT j.id FROM jobs j
			JOIN companies c ON c.id = j.company_id
			WHERE c.slug IN (%s)
		)`, slugPlaceholders), slugArgs...); err != nil {
		return fmt.Errorf("wipe job_tags: %w", err)
	}
	if _, err := db.Exec(ctx, fmt.Sprintf(`
		DELETE FROM jobs WHERE company_id IN (
			SELECT id FROM companies WHERE slug IN (%s)
		)`, slugPlaceholders), slugArgs...); err != nil {
		return fmt.Errorf("wipe jobs: %w", err)
	}
	if _, err := db.Exec(ctx, fmt.Sprintf(`
		DELETE FROM companies WHERE slug IN (%s)`, slugPlaceholders), slugArgs...); err != nil {
		return fmt.Errorf("wipe companies: %w", err)
	}
	if _, err := db.Exec(ctx, fmt.Sprintf(`
		DELETE FROM users WHERE email IN (%s)`, emailPlaceholders), emailArgs...); err != nil {
		return fmt.Errorf("wipe users: %w", err)
	}

	// ── Companies ─────────────────────────────────────────────────────────
	acme, err := insertCompany(ctx, db,
		"acme@example.com", "Acme Corp", "acme-corp",
		"We build scalable developer tooling and infrastructure for teams shipping fast.",
		"https://acme.example.com",
	)
	if err != nil {
		return err
	}

	devstudio, err := insertCompany(ctx, db,
		"devstudio@example.com", "DevStudio", "devstudio",
		"A fully remote product studio building developer-first SaaS products.",
		"https://devstudio.example.com",
	)
	if err != nil {
		return err
	}

	finledger, err := insertCompany(ctx, db,
		"finledger@example.com", "FinLedger", "finledger",
		"Next-generation financial infrastructure for modern fintech companies.",
		"https://finledger.example.com",
	)
	if err != nil {
		return err
	}

	neural, err := insertCompany(ctx, db,
		"neural-labs@example.com", "Neural Labs", "neural-labs",
		"Applied AI research lab building production ML systems at scale.",
		"https://neurallabs.example.com",
	)
	if err != nil {
		return err
	}

	forge, err := insertCompany(ctx, db,
		"forge-tools@example.com", "Forge Tools", "forge-tools",
		"Open-source dev tools company — we build the picks and shovels.",
		"https://forgetools.example.com",
	)
	if err != nil {
		return err
	}

	pixel, err := insertCompany(ctx, db,
		"pixel-agency@example.com", "Pixel Agency", "pixel-agency",
		"Design-led digital agency creating products for ambitious startups.",
		"https://pixelagency.example.com",
	)
	if err != nil {
		return err
	}

	cloudnine, err := insertCompany(ctx, db,
		"cloudnine@example.com", "CloudNine", "cloudnine",
		"Cloud infrastructure and DevOps platform trusted by 500+ engineering teams.",
		"https://cloudnine.example.com",
	)
	if err != nil {
		return err
	}

	// ── Jobs ──────────────────────────────────────────────────────────────
	type job struct {
		companyID   string
		title       string
		description string
		location    string
		salaryMin   int
		salaryMax   int
		status      string
		tags        []string
	}

	desc := func(role, detail string) string {
		return fmt.Sprintf("We're looking for a %s to join our growing team.\n\n%s\n\nRequirements:\n- Strong communication skills\n- Bias toward shipping\n- Comfortable working in a fast-paced remote environment", role, detail)
	}

	jobs := []job{
		// ── Acme Corp ──────────────────────────────────────────────────────
		{acme.id, "Senior Go Engineer", desc("senior Go engineer", "Build and own core backend services handling millions of requests per day. You'll work on distributed systems, design APIs, and mentor junior engineers."), "Remote", 100000, 140000, "open", []string{"Go", "PostgreSQL", "gRPC", "Distributed Systems"}},
		{acme.id, "Frontend Engineer (React)", desc("frontend engineer", "Build the interfaces developers use every day. React + TypeScript, collaborating directly with designers and backend engineers. We ship often."), "Remote / NYC", 90000, 120000, "open", []string{"React", "TypeScript", "CSS", "Vite"}},
		{acme.id, "Platform Engineer", desc("platform engineer", "Own our Kubernetes-based platform. Build internal tooling, improve CI/CD pipelines, and drive reliability across our stack."), "Remote", 110000, 150000, "open", []string{"Kubernetes", "Terraform", "Go", "Prometheus"}},
		{acme.id, "Engineering Manager", desc("engineering manager", "Lead a team of 5 backend engineers. Run planning, unblock your team, grow engineers, and partner with product on roadmap."), "Remote", 140000, 180000, "draft", []string{"Leadership", "Go", "Backend"}},
		{acme.id, "Security Engineer", desc("security engineer", "Own application and infrastructure security. Threat modeling, pen testing, and building secure-by-default patterns across the codebase."), "Remote", 120000, 160000, "open", []string{"Security", "Go", "AWS"}},
		{acme.id, "Staff Engineer", desc("staff engineer", "Drive technical strategy across teams. You'll set the bar for engineering quality, lead cross-team initiatives, and solve the hardest problems."), "Remote", 160000, 200000, "open", []string{"Go", "Distributed Systems", "Leadership"}},
		{acme.id, "Data Engineer", desc("data engineer", "Build and maintain our data platform. Design pipelines, own the warehouse, and enable the analytics that drive product decisions."), "Remote", 95000, 130000, "open", []string{"Data Engineering", "Python", "PostgreSQL", "AWS"}},

		// ── DevStudio ──────────────────────────────────────────────────────
		{devstudio.id, "Full-Stack Engineer", desc("full-stack engineer", "Work across Go backend and React frontend. Own features end-to-end and talk to users. Small team, huge ownership."), "Remote", 85000, 115000, "open", []string{"Go", "React", "PostgreSQL", "Full-Stack"}},
		{devstudio.id, "Product Designer", desc("product designer", "Shape the look and feel of a developer tool. Own design end-to-end: research, wireframes, high-fidelity, and implementation collaboration."), "Remote", 80000, 110000, "open", []string{"Figma", "UX", "Product Design"}},
		{devstudio.id, "DevRel Engineer", desc("devrel engineer", "Write tutorials, build sample apps, run workshops, and act as the voice of our developer community internally."), "Remote", 90000, 120000, "closed", []string{"Developer Relations", "TypeScript", "Community"}},
		{devstudio.id, "iOS Engineer", desc("iOS engineer", "Build our native iOS app from the ground up. You'll own the mobile experience and work closely with our backend team."), "Remote", 95000, 125000, "open", []string{"iOS", "Swift"}},
		{devstudio.id, "Android Engineer", desc("android engineer", "Build our native Android app. Kotlin, Jetpack Compose, and a team that values craftsmanship."), "Remote", 90000, 120000, "open", []string{"Android", "Kotlin"}},
		{devstudio.id, "Backend Engineer (Node)", desc("backend engineer", "Build APIs and integrations in Node.js + TypeScript. You'll work on our webhook infrastructure and third-party integrations."), "Remote / NYC", 80000, 110000, "draft", []string{"TypeScript", "Node.js", "PostgreSQL", "Redis"}},

		// ── FinLedger ──────────────────────────────────────────────────────
		{finledger.id, "Backend Engineer (Go)", desc("backend engineer", "Build the financial infrastructure layer. Ledger systems, double-entry bookkeeping, and payment processing at scale."), "New York", 110000, 150000, "open", []string{"Go", "PostgreSQL", "Distributed Systems"}},
		{finledger.id, "Senior Frontend Engineer", desc("senior frontend engineer", "Build our web app used by thousands of finance teams. React, TypeScript, and complex data visualization."), "New York", 105000, 140000, "open", []string{"React", "TypeScript", "GraphQL"}},
		{finledger.id, "Infrastructure Engineer", desc("infrastructure engineer", "Design and run our AWS-based infrastructure. High availability, disaster recovery, and SOC2 compliance."), "Remote", 115000, 155000, "open", []string{"AWS", "Terraform", "Kubernetes", "DevOps"}},
		{finledger.id, "Security & Compliance Engineer", desc("security engineer", "Own our security posture and compliance programs (SOC2, PCI). Pen testing, access control, and audit logging."), "New York", 120000, 160000, "open", []string{"Security", "AWS", "Compliance"}},
		{finledger.id, "Data Analyst", desc("data analyst", "Turn financial data into product insights. SQL-heavy role working alongside engineering and product teams."), "Hybrid (NYC)", 75000, 100000, "open", []string{"Python", "PostgreSQL", "Data Engineering"}},
		{finledger.id, "Rust Engineer", desc("rust engineer", "Build our high-performance transaction processing core in Rust. Latency and correctness are non-negotiable."), "Remote", 130000, 170000, "open", []string{"Rust", "Distributed Systems", "PostgreSQL"}},
		{finledger.id, "Mobile Engineer (React Native)", desc("mobile engineer", "Build our React Native app for iOS and Android. Financial data at users' fingertips."), "Remote", 90000, 120000, "draft", []string{"React", "TypeScript", "iOS", "Android"}},

		// ── Neural Labs ────────────────────────────────────────────────────
		{neural.id, "ML Engineer", desc("ML engineer", "Train and deploy production ML models. You'll work on recommendation systems, anomaly detection, and NLP pipelines."), "San Francisco", 140000, 190000, "open", []string{"Machine Learning", "Python", "AWS"}},
		{neural.id, "Research Engineer", desc("research engineer", "Bridge research and production. Implement papers, run experiments, and push our models from prototype to product."), "San Francisco", 150000, 200000, "open", []string{"Machine Learning", "Python", "Distributed Systems"}},
		{neural.id, "Data Engineer", desc("data engineer", "Build the data pipelines that feed our models. Petabyte-scale data, real-time streaming, and clean feature stores."), "Remote", 110000, 150000, "open", []string{"Data Engineering", "Python", "AWS", "Kubernetes"}},
		{neural.id, "Backend Engineer (Python)", desc("backend engineer", "Build APIs and microservices that serve our ML models in production. FastAPI, Docker, and a focus on reliability."), "Remote", 100000, 135000, "open", []string{"Python", "Docker", "Kubernetes", "PostgreSQL"}},
		{neural.id, "Platform Engineer", desc("platform engineer", "Build and own our ML platform. Training infrastructure, model serving, and experiment tracking at scale."), "Remote", 120000, 160000, "open", []string{"Kubernetes", "Python", "AWS", "Docker"}},
		{neural.id, "Frontend Engineer", desc("frontend engineer", "Build the interfaces our data scientists and customers use to interact with our models. React + TypeScript."), "Remote", 90000, 120000, "open", []string{"React", "TypeScript", "GraphQL"}},
		{neural.id, "Head of Engineering", desc("head of engineering", "Lead a team of 20 engineers across ML, backend, and infrastructure. Partner with the CEO on technical strategy."), "San Francisco", 180000, 230000, "draft", []string{"Leadership", "Machine Learning", "Distributed Systems"}},

		// ── Forge Tools ────────────────────────────────────────────────────
		{forge.id, "Open Source Engineer (Go)", desc("Go engineer", "Build and maintain our flagship open-source CLI tools. Community engagement, feature development, and performance work."), "Remote", 100000, 135000, "open", []string{"Go", "CLI", "Developer Relations"}},
		{forge.id, "Open Source Engineer (Rust)", desc("Rust engineer", "Contribute to our Rust-based build tooling. Performance, correctness, and a great developer experience."), "Remote", 110000, 145000, "open", []string{"Rust", "CLI", "DevOps"}},
		{forge.id, "Developer Experience Engineer", desc("DX engineer", "Make our tools a delight to use. Own documentation, SDKs, CLI ergonomics, and the getting-started experience."), "Remote", 95000, 130000, "open", []string{"Go", "TypeScript", "Developer Relations"}},
		{forge.id, "Full-Stack Engineer", desc("full-stack engineer", "Build the web dashboard for our tools. Go backend, Next.js frontend, and a deep care for developer UX."), "Remote", 90000, 120000, "open", []string{"Go", "Next.js", "TypeScript", "PostgreSQL"}},
		{forge.id, "DevOps Engineer", desc("DevOps engineer", "Own our internal infrastructure and help customers integrate our tools into their CI/CD pipelines."), "Remote", 100000, 135000, "open", []string{"DevOps", "Kubernetes", "Docker", "AWS"}},
		{forge.id, "Technical Writer", desc("technical writer", "Write world-class documentation, tutorials, and API references. The best tool is useless without great docs."), "Remote", 70000, 95000, "open", []string{"Developer Relations", "Community"}},
		{forge.id, "Backend Engineer (Python SDK)", desc("backend engineer", "Build and maintain our Python SDK. API design, testing, and ensuring a great experience for our Python users."), "Remote", 90000, 120000, "closed", []string{"Python", "Developer Relations", "PostgreSQL"}},

		// ── Pixel Agency ───────────────────────────────────────────────────
		{pixel.id, "Senior Product Designer", desc("senior product designer", "Lead design on 2-3 concurrent client projects. From discovery and research through to final polish and handoff."), "London", 80000, 110000, "open", []string{"Figma", "UX", "Product Design"}},
		{pixel.id, "UI Engineer", desc("UI engineer", "Implement pixel-perfect designs in React. Work closely with designers and own the frontend quality bar."), "London", 75000, 100000, "open", []string{"React", "TypeScript", "CSS"}},
		{pixel.id, "Motion Designer", desc("motion designer", "Create animations and micro-interactions that make our clients' products feel alive. After Effects, Lottie, and CSS."), "Remote", 65000, 90000, "open", []string{"Figma", "UX"}},
		{pixel.id, "Creative Director", desc("creative director", "Set the visual direction for client projects. You'll lead pitches, guide designers, and maintain a strong aesthetic point of view."), "London", 100000, 140000, "draft", []string{"Figma", "Product Design", "Leadership"}},
		{pixel.id, "Frontend Engineer (Next.js)", desc("frontend engineer", "Build marketing sites and web apps for our clients. Next.js, TypeScript, and a focus on performance and accessibility."), "Remote", 80000, 110000, "open", []string{"Next.js", "TypeScript", "React", "CSS"}},
		{pixel.id, "Account Manager", desc("account manager", "Manage relationships with 8-10 client accounts. You'll own project scoping, timelines, and client satisfaction."), "London", 60000, 85000, "closed", []string{"Community", "Leadership"}},

		// ── CloudNine ──────────────────────────────────────────────────────
		{cloudnine.id, "Site Reliability Engineer", desc("SRE", "Own the reliability of our platform. On-call, incident response, chaos engineering, and building automation to reduce toil."), "Remote", 120000, 160000, "open", []string{"Kubernetes", "Go", "Prometheus", "AWS"}},
		{cloudnine.id, "Backend Engineer (Go)", desc("backend engineer", "Build the control plane of our cloud platform. APIs, scheduling, resource management, and a deep focus on reliability."), "Remote", 110000, 150000, "open", []string{"Go", "Kubernetes", "gRPC", "PostgreSQL"}},
		{cloudnine.id, "Infrastructure Engineer", desc("infrastructure engineer", "Own the physical and virtual infrastructure our platform runs on. Networking, storage, and multi-region deployments."), "Remote", 115000, 155000, "open", []string{"AWS", "Terraform", "DevOps", "Kubernetes"}},
		{cloudnine.id, "Security Engineer", desc("security engineer", "Secure our multi-tenant cloud platform. Zero-trust networking, secrets management, and customer isolation."), "Remote", 125000, 165000, "open", []string{"Security", "Kubernetes", "AWS"}},
		{cloudnine.id, "Frontend Engineer", desc("frontend engineer", "Build the dashboard our customers use to manage their infrastructure. React, TypeScript, and complex real-time data."), "Remote", 95000, 130000, "open", []string{"React", "TypeScript", "GraphQL"}},
		{cloudnine.id, "CLI Engineer (Go)", desc("Go engineer", "Build and maintain our CLI — the primary interface for thousands of engineers. Ergonomics, speed, and great error messages."), "Remote", 100000, 135000, "open", []string{"Go", "CLI", "Developer Relations"}},
		{cloudnine.id, "Billing Engineer", desc("billing engineer", "Own our usage-based billing infrastructure. Metering, invoicing, and integrations with Stripe and our cloud provider APIs."), "Austin", 105000, 140000, "open", []string{"Go", "PostgreSQL", "AWS"}},
		{cloudnine.id, "Technical Account Manager", desc("technical account manager", "Help our enterprise customers succeed. You'll run onboarding, provide technical guidance, and be their internal advocate."), "Austin", 85000, 115000, "draft", []string{"DevOps", "Community", "Leadership"}},
		{cloudnine.id, "Data Engineer", desc("data engineer", "Build the pipelines for our usage metrics and billing data. Petabytes of telemetry, clean and on time."), "Remote", 100000, 135000, "open", []string{"Data Engineering", "Python", "AWS", "PostgreSQL"}},
		{cloudnine.id, "Senior Go Engineer", desc("senior Go engineer", "Architect and build new platform features. You'll own large surface areas and mentor the engineers around you."), "Berlin", 110000, 150000, "open", []string{"Go", "Kubernetes", "Distributed Systems"}},
		{cloudnine.id, "Rust Systems Engineer", desc("Rust engineer", "Build the hot path of our data plane in Rust. Microsecond latency, zero-copy networking, and bulletproof correctness."), "Remote", 135000, 180000, "open", []string{"Rust", "Distributed Systems", "DevOps"}},
	}

	for _, j := range jobs {
		if err := insertJob(ctx, db, j.companyID, j.title, j.description, j.location, j.salaryMin, j.salaryMax, j.status, j.tags); err != nil {
			return err
		}
	}

	// ── Seeker user ───────────────────────────────────────────────────────
	var seekerID string
	if err := db.QueryRow(ctx, `
		INSERT INTO users (email, password_hash, role)
		VALUES ($1, $2, 'seeker') RETURNING id
	`, "seeker@example.com", hash("password123")).Scan(&seekerID); err != nil {
		return fmt.Errorf("insert seeker: %w", err)
	}

	// Apply to first open Acme job
	var firstJobID string
	if err := db.QueryRow(ctx, `
		SELECT j.id FROM jobs j
		JOIN companies c ON c.id = j.company_id
		WHERE j.status = 'open' AND c.slug = 'acme-corp'
		ORDER BY j.created_at LIMIT 1
	`).Scan(&firstJobID); err != nil {
		return fmt.Errorf("find first open job: %w", err)
	}
	if _, err := db.Exec(ctx, `
		INSERT INTO applications (job_id, user_id, cover_note) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING
	`, firstJobID, seekerID, "I have 5 years of Go experience and have built distributed systems at scale. Excited about this role."); err != nil {
		return fmt.Errorf("insert application: %w", err)
	}

	total := len(jobs)
	open := 0
	for _, j := range jobs {
		if j.status == "open" {
			open++
		}
	}

	fmt.Printf("\n%d jobs inserted (%d open)\n", total, open)
	fmt.Println("\nSeed accounts (password: password123)")
	fmt.Println("  company : acme@example.com")
	fmt.Println("  company : devstudio@example.com")
	fmt.Println("  company : finledger@example.com")
	fmt.Println("  company : neural-labs@example.com")
	fmt.Println("  company : forge-tools@example.com")
	fmt.Println("  company : pixel-agency@example.com")
	fmt.Println("  company : cloudnine@example.com")
	fmt.Println("  seeker  : seeker@example.com")

	return nil
}
