package main

import "time"

const (
	STORAGE_BASE         = "./data"
	TRANSCRIPTS_DIR      = STORAGE_BASE + "/transcripts"
	ANALYSIS_DIR         = STORAGE_BASE + "/analysis"
	AGGREGATES_DIR       = STORAGE_BASE + "/aggregates"
	TICKETS_DIR          = STORAGE_BASE + "/tickets"
	AGGREGATION_INTERVAL = 1 * time.Minute // for dev. In prod set to 24h.
	SERVER_LISTEN_ADDR   = ":8080"
)

// Feature buckets for problem categorization
var FeatureBuckets = []string{
	"Lead Management",
	"Lead Quality",
	"Lead Quantity",
	"Promoted Listing / Lead Priority",
	"Visibility / Ranking",
	"TrustSEAL / Verification",
	"Catalog / Storefront Setup",
	"Buyer Interaction",
	"BizInsight Analytics",
	"Billing & Renewal",
	"Payments",
	"App / Platform Usability",
	"Support / Training",
	"Seller Verification",
	"Compliance / Documentation",
	"Category-City Targeting",
	"Communication",
	"Account / Dashboard",
	"Other",
}

// IndiaMART Business Context - Comprehensive knowledge base for AI analysis
const IndiaMARTContext = `
=== INDIAMART BUSINESS OVERVIEW ===

IndiaMART is India's LARGEST B2B online marketplace, founded in 1996 by Dinesh Chandra Agarwal.
- Mission: "Make doing business easy" and democratize business opportunities for all
- Platform connects 86+ lakh (8.6 million) suppliers with 21.9+ crore (219 million) registered buyers
- Headquarters: Noida, India (Listed company: IndiaMART InterMESH Ltd.)
- Focus: Digital and financial inclusion for MSMEs (Micro, Small & Medium Enterprises)

=== PAID SUBSCRIPTION PRODUCTS (Most Important for Call Analysis) ===

1. MDC (Mini Dynamic Catalogue) - Entry Level
   - Price: Rs.4,000/month | Rs.35,000/year | Rs.60,000/2yr | Rs.72,000/3yr
   - Benefits:
     * 7 weekly BuyLeads (monthly plan) / 10 weekly (yearly plan) + 1 daily bonus
     * Higher listing vs free suppliers
     * Lead Management System (LMS) with instant alerts
     * Zero missed calls - Call forwarding (PNS) to up to 5 numbers
     * Professional catalog designed by experts

2. TrustSEAL - Trust & Verification Badge
   - Price: Rs.50,000/year | Rs.80,000/2yr | Rs.100,000/3yr
   - Benefits:
     * TrustSEAL logo and stamp on catalog (credibility badge)
     * TrustSEAL certificate (physical + e-copy)
     * 20 domestic BuyLeads/week + 2 daily bonus BuyLeads
     * Higher listing than MDC suppliers
     * Document verification by IndiaMART team
   - Common Issues: Badge not displaying, verification delays, document pending

3. Maximiser - Premium Catalog Service
   - Price: Rs.75,000/year | Rs.1,20,000/2yr | Rs.1,50,000/3yr
   - Benefits:
     * Personal domain (.com/.net/.in/.co.in) with server access
     * 4 corporate email IDs
     * Up to 10,000 products in catalog
     * TrustSEAL badge included
     * 30 weekly domestic BuyLeads + 2 daily bonus
     * Mobile-responsive templates
     * PDF catalog with 360-degree visibility

4. IM Star Pro - Premium Visibility
   - Benefits:
     * Star Supplier label/badge
     * Dynamic cities visibility (behavior-based districts)
     * Premium listing in dynamic cities & preferred locations
     * NO limitation on BuyLeads - unlimited leads
     * Higher visibility in local cities and chosen categories

5. IM Leader Pro - Top Tier
   - Benefits:
     * Leading Supplier label
     * AI-based targeting for business acceleration
     * Premium listing with highest visibility
     * Dynamic cities + preferred locations
     * Unlimited BuyLeads
     * Top visibility in local cities and categories

=== KEY FEATURES & TOOLS FOR SELLERS ===

1. BuyLeads System
   - BuyLeads are buyer inquiries/requirements posted by buyers
   - Sellers receive BuyLeads based on their subscription
   - Types: Daily bonus BuyLeads, Weekly domestic BuyLeads
   - Quality concern: Fake/spam inquiries from competitors or students

2. Lead Management System (LMS)
   - 24x7 instant alerts for new leads
   - Smart lead tracking and management
   - Contact buyer directly via call/message

3. Preferred Number Service (PNS)
   - Call forwarding to up to 5 numbers
   - Zero missed calls feature
   - Available in paid plans

4. Catalog/Storefront Features
   - Product listing with images, prices, descriptions
   - Response rate tracking (affects ranking)
   - Catalog optimization by IndiaMART team
   - Last updated date visible (affects algorithm)

5. Ranking Algorithm Factors
   - Response rate to inquiries (CRITICAL - should be >80%)
   - Catalog update frequency
   - Subscription level (paid > free)
   - TrustSEAL/Star/Leader badge
   - Business verification status

6. BizInsight Analytics
   - Dashboard showing business metrics
   - Lead conversion tracking
   - Visibility reports

=== COMMON SELLER ISSUES (Categorize calls into these) ===

1. Lead Quality Issues
   - Fake inquiries from competitors
   - Students doing market research
   - Non-serious buyers
   - Spam leads affecting response rate

2. Lead Quantity Issues
   - Not receiving promised number of leads
   - Leads stopped coming
   - Daily/weekly quota not met

3. Visibility/Ranking Issues
   - Products not on first page
   - Dropped from top rankings
   - Competitors showing above despite same/lower subscription

4. TrustSEAL/Verification Issues
   - Badge not displaying after payment
   - Document verification pending/delayed
   - Certificate not received
   - Badge removed unexpectedly

5. Billing & Renewal Issues
   - Payment charged but subscription not activated
   - Renewal pricing disputes
   - Refund requests
   - Auto-renewal complaints
   - Package downgrade issues

6. Technical/Platform Issues
   - App not working
   - Dashboard errors
   - Leads not syncing
   - Unable to respond to inquiries
   - Login problems

7. Catalog Issues
   - Products not uploaded
   - Images not displaying
   - Category mismatch
   - Catalog not updated by team

8. Communication Issues
   - Multiple calls with no resolution
   - Promised callback not received
   - Escalation not happening
   - Agent commitments not honored

=== COMPETITOR CONTEXT ===
Main competitors sellers may threaten to switch to:
- TradeIndia
- JustDial
- Amazon Business
- Alibaba
- IndiaBizClub

=== SELLER PSYCHOLOGY & BUSINESS IMPACT ===

Understand these seller concerns:
1. ROI Focus: Sellers pay substantial amounts (Rs.35K-1.5L/year) and expect measurable returns
2. Lead Conversion: Quality leads = actual orders = business survival for MSMEs
3. Trust: TrustSEAL is identity/credibility marker - its absence affects buyer trust
4. Competition: Visibility directly impacts business - ranking drop = lost revenue
5. Cash Flow: Small businesses have tight margins - billing issues are critical
6. Time Sensitivity: Business inquiries need quick response - delays = lost deals

=== AGENT PERFORMANCE MARKERS ===

Good Agent Behavior:
- Acknowledges problem genuinely
- Checks account/system before responding
- Provides specific solutions, not generic advice
- Takes ownership with follow-up commitment
- Offers compensation/credit when appropriate
- Escalates when needed

Poor Agent Behavior:
- Generic responses without checking account
- Blaming seller (update catalog, improve response rate)
- No ownership or follow-up commitment
- Multiple transfers without resolution
- Ignoring escalation requests
- Making promises without capability to deliver

=== CHURN RISK INDICATORS ===

High Churn Risk:
- Mentions competitor names (TradeIndia, JustDial)
- Multiple unresolved complaints
- Says "last chance" or "final warning"
- Threatens cancellation/refund
- Been a customer for 2+ years with recent issues
- Premium customer (TrustSEAL/Maximiser) with problems

Low Churn Risk:
- New customer learning the platform
- Single minor issue
- Expresses satisfaction with resolution
- Asks about upgrading services
- Shows interest in additional products

=== UPSELL OPPORTUNITIES ===

Look for these signals:
- Seller mentions competitor's features they want
- Asks about premium services
- Wants more leads/visibility
- Business is growing
- Interested in website/domain
- Needs better credibility (→ TrustSEAL)
- Wants unlimited leads (→ IM Star/Leader Pro)
`
