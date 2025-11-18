package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// HeroSlide represents the hero carousel cards rendered on the landing page
// It mirrors the shape the frontend HeroContent component expects.
type HeroSlide struct {
	ID          primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	Title       string              `bson:"title" json:"title"`
	Subtitle    string              `bson:"subtitle" json:"subtitle"`
	Price       string              `bson:"price" json:"price"` // Display price (e.g., "₹450")
	Description string              `bson:"description" json:"description"`
	Image       string              `bson:"image" json:"image"`
	Features    []string            `bson:"features" json:"features"`
	Gradient    string              `bson:"gradient" json:"gradient"`
	GlowColor   string              `bson:"glowColor" json:"glowColor"`
	Position    int                 `bson:"position" json:"position"`
	ProductID   *primitive.ObjectID `bson:"productId,omitempty" json:"productId,omitempty"`
	Product     *Product            `bson:"product,omitempty" json:"product,omitempty"`
	// Product fields for creating/updating in products collection
	Brand              string     `bson:"brand,omitempty" json:"brand,omitempty"`
	ProductPrice       float64    `bson:"productPrice,omitempty" json:"productPrice,omitempty"` // Actual numeric price
	Category           string     `bson:"category,omitempty" json:"category,omitempty"`
	MainCategory       string     `bson:"mainCategory,omitempty" json:"mainCategory,omitempty"`
	Subcategory        string     `bson:"subcategory,omitempty" json:"subcategory,omitempty"`
	Images             []string   `bson:"images,omitempty" json:"images,omitempty"`
	Stock              int        `bson:"stock,omitempty" json:"stock,omitempty"`
	Gender             string     `bson:"gender,omitempty" json:"gender,omitempty"`
	DialColor          string     `bson:"dialColor,omitempty" json:"dialColor,omitempty"`
	DialShape          string     `bson:"dialShape,omitempty" json:"dialShape,omitempty"`
	DialType           string     `bson:"dialType,omitempty" json:"dialType,omitempty"`
	StrapColor         string     `bson:"strapColor,omitempty" json:"strapColor,omitempty"`
	StrapMaterial      string     `bson:"strapMaterial,omitempty" json:"strapMaterial,omitempty"`
	Style              string     `bson:"style,omitempty" json:"style,omitempty"`
	DialThickness      string     `bson:"dialThickness,omitempty" json:"dialThickness,omitempty"`
	DiscountPercentage *float64   `bson:"discountPercentage,omitempty" json:"discountPercentage,omitempty"`
	DiscountAmount     *float64   `bson:"discountAmount,omitempty" json:"discountAmount,omitempty"`
	DiscountStartDate  *time.Time `bson:"discountStartDate,omitempty" json:"discountStartDate,omitempty"`
	DiscountEndDate    *time.Time `bson:"discountEndDate,omitempty" json:"discountEndDate,omitempty"`
	CreatedAt          time.Time  `bson:"createdAt" json:"createdAt"`
	UpdatedAt          time.Time  `bson:"updatedAt" json:"updatedAt"`
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
	ID           primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	Tagline      string              `bson:"tagline" json:"tagline"`
	Title        string              `bson:"title" json:"title"`
	Description  string              `bson:"description" json:"description"`
	Price        string              `bson:"price,omitempty" json:"price,omitempty"` // Display price
	Availability string              `bson:"availability" json:"availability"`
	CtaLabel     string              `bson:"ctaLabel" json:"ctaLabel"`
	CtaHref      string              `bson:"ctaHref" json:"ctaHref"`
	Image        string              `bson:"image" json:"image"`
	ImageAlt     string              `bson:"imageAlt" json:"imageAlt"`
	Layout       string              `bson:"layout" json:"layout"`
	Position     int                 `bson:"position" json:"position"`
	ProductID    *primitive.ObjectID `bson:"productId,omitempty" json:"productId,omitempty"`
	Product      *Product            `bson:"product,omitempty" json:"product,omitempty"`
	// Product fields for creating/updating in products collection
	Brand              string     `bson:"brand,omitempty" json:"brand,omitempty"`
	ProductPrice       float64    `bson:"productPrice,omitempty" json:"productPrice,omitempty"` // Actual numeric price
	Category           string     `bson:"category,omitempty" json:"category,omitempty"`
	MainCategory       string     `bson:"mainCategory,omitempty" json:"mainCategory,omitempty"`
	Subcategory        string     `bson:"subcategory,omitempty" json:"subcategory,omitempty"`
	Images             []string   `bson:"images,omitempty" json:"images,omitempty"`
	Stock              int        `bson:"stock,omitempty" json:"stock,omitempty"`
	Gender             string     `bson:"gender,omitempty" json:"gender,omitempty"`
	DialColor          string     `bson:"dialColor,omitempty" json:"dialColor,omitempty"`
	DialShape          string     `bson:"dialShape,omitempty" json:"dialShape,omitempty"`
	DialType           string     `bson:"dialType,omitempty" json:"dialType,omitempty"`
	StrapColor         string     `bson:"strapColor,omitempty" json:"strapColor,omitempty"`
	StrapMaterial      string     `bson:"strapMaterial,omitempty" json:"strapMaterial,omitempty"`
	Style              string     `bson:"style,omitempty" json:"style,omitempty"`
	DialThickness      string     `bson:"dialThickness,omitempty" json:"dialThickness,omitempty"`
	DiscountPercentage *float64   `bson:"discountPercentage,omitempty" json:"discountPercentage,omitempty"`
	DiscountAmount     *float64   `bson:"discountAmount,omitempty" json:"discountAmount,omitempty"`
	DiscountStartDate  *time.Time `bson:"discountStartDate,omitempty" json:"discountStartDate,omitempty"`
	DiscountEndDate    *time.Time `bson:"discountEndDate,omitempty" json:"discountEndDate,omitempty"`
	CreatedAt          time.Time  `bson:"createdAt" json:"createdAt"`
	UpdatedAt          time.Time  `bson:"updatedAt" json:"updatedAt"`
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

// GalleryImage represents a single image in the homepage gallery section
type GalleryImage struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Url       string             `bson:"url" json:"url"`
	Alt       string             `bson:"alt" json:"alt"`
	Position  int                `bson:"position" json:"position"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time          `bson:"updatedAt" json:"updatedAt"`
}

// Extend HomeContent to include gallery images (backwards compatible for existing clients not using it)
type HomeContentWithGallery struct {
	HomeContent
	Gallery []GalleryImage `json:"gallery"`
}
