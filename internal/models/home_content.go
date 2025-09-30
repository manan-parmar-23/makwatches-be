package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// HeroSlide represents the hero carousel cards rendered on the landing page
// It mirrors the shape the frontend HeroContent component expects.
type HeroSlide struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Title       string             `bson:"title" json:"title"`
	Subtitle    string             `bson:"subtitle" json:"subtitle"`
	Price       string             `bson:"price" json:"price"`
	Description string             `bson:"description" json:"description"`
	Image       string             `bson:"image" json:"image"`
	Features    []string           `bson:"features" json:"features"`
	Gradient    string             `bson:"gradient" json:"gradient"`
	GlowColor   string             `bson:"glowColor" json:"glowColor"`
	Position    int                `bson:"position" json:"position"`
	CreatedAt   time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time          `bson:"updatedAt" json:"updatedAt"`
}

// HomeCategoryCard powers the curated category tiles on the landing page.
type HomeCategoryCard struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Title      string             `bson:"title" json:"title"`
	Subtitle   string             `bson:"subtitle" json:"subtitle"`
	Href       string             `bson:"href" json:"href"`
	Image      string             `bson:"image" json:"image"`
	BgGradient string             `bson:"bgGradient" json:"bgGradient"`
	Position   int                `bson:"position" json:"position"`
	CreatedAt  time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt  time.Time          `bson:"updatedAt" json:"updatedAt"`
}

// HomeCollectionFeature represents the collection spotlight sections.
type HomeCollectionFeature struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Tagline      string             `bson:"tagline" json:"tagline"`
	Title        string             `bson:"title" json:"title"`
	Description  string             `bson:"description" json:"description"`
	Availability string             `bson:"availability" json:"availability"`
	CtaLabel     string             `bson:"ctaLabel" json:"ctaLabel"`
	CtaHref      string             `bson:"ctaHref" json:"ctaHref"`
	Image        string             `bson:"image" json:"image"`
	ImageAlt     string             `bson:"imageAlt" json:"imageAlt"`
	Layout       string             `bson:"layout" json:"layout"`
	Position     int                `bson:"position" json:"position"`
	CreatedAt    time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt    time.Time          `bson:"updatedAt" json:"updatedAt"`
}

// TechShowcaseHighlight controls the short highlight banner in the tech showcase section.
type TechShowcaseHighlight struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Value      string             `bson:"value" json:"value"`
	Title      string             `bson:"title" json:"title"`
	Subtitle   string             `bson:"subtitle" json:"subtitle"`
	AccentHex  string             `bson:"accentHex" json:"accentHex"`
	Background string             `bson:"background" json:"background"`
	CreatedAt  time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt  time.Time          `bson:"updatedAt" json:"updatedAt"`
}

// TechShowcaseCard represents the cards rendered inside the tech showcase grid.
type TechShowcaseCard struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Title           string             `bson:"title" json:"title"`
	Subtitle        string             `bson:"subtitle" json:"subtitle"`
	Image           string             `bson:"image" json:"image"`
	BackgroundImage string             `bson:"backgroundImage" json:"backgroundImage"`
	Rating          float64            `bson:"rating" json:"rating"`
	ReviewCount     int                `bson:"reviewCount" json:"reviewCount"`
	Badge           string             `bson:"badge" json:"badge"`
	Color           string             `bson:"color" json:"color"`
	Position        int                `bson:"position" json:"position"`
	CreatedAt       time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt       time.Time          `bson:"updatedAt" json:"updatedAt"`
}

// HomeContent bundles all landing page sections for the storefront response.
type HomeContent struct {
	HeroSlides  []HeroSlide             `json:"heroSlides"`
	Categories  []HomeCategoryCard      `json:"categories"`
	Collections []HomeCollectionFeature `json:"collections"`
	TechCards   []TechShowcaseCard      `json:"techCards"`
	Highlight   *TechShowcaseHighlight  `json:"highlight"`
}
