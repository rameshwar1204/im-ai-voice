# IndiaMART Voice AI - Call Transcript Analysis System

## 📋 Table of Contents
1. [Overview](#overview)
2. [System Architecture](#system-architecture)
3. [Flow Diagrams](#flow-diagrams)
4. [Core Components](#core-components)
5. [Data Models](#data-models)
6. [API Endpoints](#api-endpoints)
7. [Dashboard UI](#dashboard-ui)
8. [Setup & Configuration](#setup--configuration)
9. [How It Works - Step by Step](#how-it-works---step-by-step)

---

## 🎯 Overview

**IndiaMART Voice AI** is an intelligent system that automatically analyzes customer support call transcripts to extract actionable business insights. It uses Google's Gemini AI to understand conversations (in Hindi/English/Hinglish) and generates:

- **Issue Detection** - Identifies problems mentioned in calls
- **Sentiment Analysis** - Understands seller mood (Positive/Neutral/Negative)
- **Churn Prediction** - Predicts likelihood of seller leaving (Low/Medium/High)
- **Upsell Opportunities** - Identifies sellers interested in more features
- **Automated Tickets** - Creates tickets for recurring issues
- **Seller Health Scores** - Tracks seller satisfaction over time

### Key Features
- ✅ **Event-Driven Processing** - New transcripts auto-analyzed
- ✅ **Real-time Dashboard** - Visual insights at a glance
- ✅ **MongoDB Storage** - Persistent data storage
- ✅ **Multi-language Support** - Hindi, English, Hinglish
- ✅ **Auto-Aggregation** - Daily summaries generated automatically
- ✅ **Ticket Generation** - Issues automatically escalated

---

## 🏗️ System Architecture

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                           IndiaMART Voice AI System                              │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                  │
│  ┌──────────────┐     ┌──────────────┐     ┌──────────────┐     ┌─────────────┐ │
│  │   Transcript │     │   Watcher    │     │   Gemini AI  │     │   MongoDB   │ │
│  │    Files     │────▶│   Service    │────▶│   Analysis   │────▶│   Storage   │ │
│  │  (*.json)    │     │              │     │              │     │             │ │
│  └──────────────┘     └──────────────┘     └──────────────┘     └─────────────┘ │
│         │                    │                    │                    │        │
│         │                    │                    │                    │        │
│         │                    ▼                    │                    │        │
│         │           ┌──────────────┐              │                    │        │
│         │           │  Aggregation │◀─────────────┘                    │        │
│         │           │   Service    │                                   │        │
│         │           └──────────────┘                                   │        │
│         │                    │                                         │        │
│         │                    ▼                                         │        │
│         │           ┌──────────────┐                                   │        │
│         │           │   Ticket     │──────────────────────────────────▶│        │
│         │           │  Generator   │                                   │        │
│         │           └──────────────┘                                   │        │
│         │                                                              │        │
│         │           ┌──────────────┐                                   │        │
│         └──────────▶│  REST API    │◀──────────────────────────────────┘        │
│                     │   Server     │                                            │
│                     └──────────────┘                                            │
│                            │                                                    │
│                            ▼                                                    │
│                     ┌──────────────┐                                            │
│                     │  Dashboard   │                                            │
│                     │     UI       │                                            │
│                     └──────────────┘                                            │
│                                                                                  │
└─────────────────────────────────────────────────────────────────────────────────┘
```

### Tech Stack
| Component | Technology |
|-----------|------------|
| Backend | Go 1.24 |
| AI Engine | Google Gemini 2.0 Flash |
| Database | MongoDB Atlas |
| Frontend | HTML5, CSS3, JavaScript |
| API | REST HTTP |

---

## 🔄 Flow Diagrams

### 1. Main Processing Flow

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                          MAIN PROCESSING FLOW                                    │
└─────────────────────────────────────────────────────────────────────────────────┘

    ┌─────────────┐
    │ New Call    │
    │ Transcript  │
    │ Arrives     │
    └──────┬──────┘
           │
           ▼
    ┌─────────────┐     No      ┌─────────────┐
    │  Already    │────────────▶│   Process   │
    │ Processed?  │             │  Transcript │
    └──────┬──────┘             └──────┬──────┘
           │ Yes                       │
           ▼                           ▼
    ┌─────────────┐             ┌─────────────┐
    │    Skip     │             │  Send to    │
    │             │             │  Gemini AI  │
    └─────────────┘             └──────┬──────┘
                                       │
                                       ▼
                                ┌─────────────┐
                                │   Parse     │
                                │  Response   │
                                └──────┬──────┘
                                       │
                     ┌─────────────────┼─────────────────┐
                     │                 │                 │
                     ▼                 ▼                 ▼
              ┌─────────────┐  ┌─────────────┐  ┌─────────────┐
              │    Save     │  │   Update    │  │   Check     │
              │  Analysis   │  │   Seller    │  │  Threshold  │
              │  to MongoDB │  │   Profile   │  │   (10?)     │
              └─────────────┘  └─────────────┘  └──────┬──────┘
                                                       │
                                          ┌───────────┴───────────┐
                                          │ Yes                   │ No
                                          ▼                       ▼
                                   ┌─────────────┐         ┌─────────────┐
                                   │    Run      │         │   Wait for  │
                                   │ Aggregation │         │   More      │
                                   └──────┬──────┘         └─────────────┘
                                          │
                                          ▼
                                   ┌─────────────┐
                                   │  Generate   │
                                   │   Tickets   │
                                   └─────────────┘
```

### 2. AI Analysis Flow

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                           AI ANALYSIS FLOW                                       │
└─────────────────────────────────────────────────────────────────────────────────┘

    ┌─────────────┐
    │ Transcript  │
    │   (Hindi/   │
    │  Hinglish)  │
    └──────┬──────┘
           │
           ▼
    ┌─────────────────────────────────────────┐
    │         BUILD PROMPT                     │
    │  ┌────────────────────────────────────┐ │
    │  │ System Prompt:                     │ │
    │  │ - IndiaMART context               │ │
    │  │ - Feature buckets                 │ │
    │  │ - Analysis guidelines             │ │
    │  └────────────────────────────────────┘ │
    │  ┌────────────────────────────────────┐ │
    │  │ User Prompt:                       │ │
    │  │ - Transcript text                 │ │
    │  │ - Expected JSON structure         │ │
    │  └────────────────────────────────────┘ │
    └──────────────────┬──────────────────────┘
                       │
                       ▼
              ┌─────────────────┐
              │   Gemini API    │
              │  (gemini-2.0-   │
              │     flash)      │
              └────────┬────────┘
                       │
                       ▼
              ┌─────────────────┐
              │  JSON Response  │
              └────────┬────────┘
                       │
                       ▼
    ┌─────────────────────────────────────────┐
    │           EXTRACTED DATA                 │
    │  ┌─────────────┐  ┌─────────────────┐   │
    │  │   Issues    │  │    Sentiment    │   │
    │  │  - Problem  │  │  - Positive     │   │
    │  │  - Bucket   │  │  - Neutral      │   │
    │  │  - Severity │  │  - Negative     │   │
    │  └─────────────┘  └─────────────────┘   │
    │  ┌─────────────┐  ┌─────────────────┐   │
    │  │ Churn Risk  │  │     Upsell      │   │
    │  │  - Low      │  │  - Score (1-10) │   │
    │  │  - Medium   │  │  - Features     │   │
    │  │  - High     │  │  - Opportunity  │   │
    │  └─────────────┘  └─────────────────┘   │
    │  ┌─────────────────────────────────┐    │
    │  │         Call Summary            │    │
    │  └─────────────────────────────────┘    │
    └─────────────────────────────────────────┘
```

### 3. Seller Profile Update Flow

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                      SELLER PROFILE UPDATE FLOW                                  │
└─────────────────────────────────────────────────────────────────────────────────┘

    ┌─────────────┐
    │    New      │
    │  Analysis   │
    └──────┬──────┘
           │
           ▼
    ┌─────────────┐     No      ┌─────────────┐
    │   Profile   │────────────▶│   Create    │
    │   Exists?   │             │   New       │
    └──────┬──────┘             │  Profile    │
           │ Yes                └──────┬──────┘
           │                           │
           ▼                           │
    ┌─────────────┐                    │
    │    Load     │                    │
    │   Profile   │                    │
    └──────┬──────┘                    │
           │                           │
           ├───────────────────────────┘
           │
           ▼
    ┌─────────────────────────────────────────┐
    │          UPDATE PROFILE                  │
    │  ┌─────────────────────────────────┐    │
    │  │  - Increment total_calls        │    │
    │  │  - Update last_call_at          │    │
    │  │  - Add to call_history          │    │
    │  │  - Update active_issues         │    │
    │  │  - Recalculate health_score     │    │
    │  │  - Adjust churn_risk            │    │
    │  └─────────────────────────────────┘    │
    └──────────────────┬──────────────────────┘
                       │
                       ▼
              ┌─────────────────┐
              │  Calculate      │
              │  Health Score   │
              │  (0-100)        │
              └────────┬────────┘
                       │
           ┌───────────┼───────────┐
           │           │           │
           ▼           ▼           ▼
    ┌──────────┐ ┌──────────┐ ┌──────────┐
    │ Critical │ │ At Risk  │ │ Healthy  │
    │  (0-39)  │ │ (40-69)  │ │ (70-100) │
    └──────────┘ └──────────┘ └──────────┘
```

### 4. Ticket Generation Flow

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                        TICKET GENERATION FLOW                                    │
└─────────────────────────────────────────────────────────────────────────────────┘

    ┌─────────────┐
    │   Daily     │
    │ Aggregate   │
    └──────┬──────┘
           │
           ▼
    ┌─────────────────────────────────────────┐
    │    GROUP ISSUES BY BUCKET               │
    │  ┌─────────────────────────────────┐    │
    │  │ Lead Quality:        15 issues  │    │
    │  │ Billing & Renewal:   12 issues  │    │
    │  │ App Usability:        8 issues  │    │
    │  │ Lead Quantity:        6 issues  │    │
    │  │ Communication:        4 issues  │    │
    │  │ ...                             │    │
    │  └─────────────────────────────────┘    │
    └──────────────────┬──────────────────────┘
                       │
                       ▼
    ┌─────────────┐     No
    │  Count ≥ 3  │────────────▶ Skip
    │   Issues?   │
    └──────┬──────┘
           │ Yes
           ▼
    ┌─────────────────────────────────────────┐
    │         CREATE TICKET                    │
    │  ┌─────────────────────────────────┐    │
    │  │ - Title: [Bucket] Issues        │    │
    │  │ - Priority: Based on severity   │    │
    │  │ - Description: Top problems     │    │
    │  │ - Affected Sellers: List        │    │
    │  │ - Status: open                  │    │
    │  └─────────────────────────────────┘    │
    └──────────────────┬──────────────────────┘
                       │
                       ▼
              ┌─────────────────┐
              │    Save to      │
              │    MongoDB      │
              └─────────────────┘
```

---

## 🧩 Core Components

### File Structure
```
im-ai-voice/
├── main.go              # Application entry point
├── config.go            # Configuration constants
├── models.go            # Data structures
├── service.go           # Business logic
├── watcher.go           # Event-driven transcript processor
├── gemini_client.go     # Google Gemini AI integration
├── mongodb.go           # MongoDB operations
├── storage.go           # File storage operations
├── router.go            # HTTP API endpoints
├── seller_profile.go    # Seller profile management
├── utils.go             # Utility functions
├── static/              # Dashboard UI
│   ├── index.html       # Main HTML
│   ├── app.js           # JavaScript logic
│   └── style.css        # Styling
└── data/
    └── transcripts/     # Input transcript files
```

### Component Descriptions

| File | Purpose |
|------|---------|
| `main.go` | Initializes all components, starts server |
| `watcher.go` | Monitors `data/transcripts/` for new files, triggers analysis |
| `gemini_client.go` | Sends transcripts to Gemini AI, parses responses |
| `service.go` | Core business logic - analysis, aggregation, tickets |
| `mongodb.go` | Database operations (save/load profiles, analyses, tickets) |
| `seller_profile.go` | Manages seller health scores and history |
| `router.go` | REST API endpoints for dashboard |

---

## 📊 Data Models

### 1. RawTranscript (Input)
```json
{
  "call_id": "667438696",
  "seller_id": "18888",
  "transcript_text": "Customer: Ji sir...",
  "customer_type": "CATALOG",
  "vintage": 24,
  "timestamp": "2025-12-12T10:30:00Z"
}
```

### 2. AnalysisResult (AI Output)
```json
{
  "call_id": "667438696",
  "seller_id": "18888",
  "call_summary": "Seller called about lead quality issues...",
  "issues": [
    {
      "problem": "Receiving irrelevant leads from other states",
      "bucket": "Lead Quality",
      "severity": "high",
      "actionable_summary": "Enable geographic filtering"
    }
  ],
  "intent": {
    "sentiment": "Negative",
    "satisfaction_score": 2
  },
  "churn": {
    "is_likely_to_churn": "high",
    "renewal_at_risk": true,
    "churn_reason": "Poor lead quality"
  },
  "upsell": {
    "has_opportunity": false,
    "score": 2
  }
}
```

### 3. SellerProfile
```json
{
  "gluser_id": "18888",
  "customer_type": "CATALOG",
  "total_calls": 5,
  "health_score": 35,
  "health_label": "Critical",
  "churn_risk": "high",
  "active_issues": ["Lead Quality", "Billing"],
  "call_history": [...]
}
```

### 4. DailyAggregate
```json
{
  "date": "2025-12-12",
  "total_calls": 84,
  "total_issues": 207,
  "feature_buckets": {
    "Lead Quality": { "total_count": 25, "affected_sellers": 12 },
    "Billing & Renewal": { "total_count": 18, "affected_sellers": 8 }
  },
  "sentiment_breakdown": { "Negative": 30, "Neutral": 45, "Positive": 9 },
  "churn_risk_breakdown": { "high": 15, "medium": 35, "low": 34 }
}
```

### 5. Ticket
```json
{
  "ticket_id": "2025-12-12-Lead_Quality-01",
  "feature_bucket": "Lead Quality",
  "priority": 1,
  "title": "Lead Quality Issues - High Priority",
  "affected_count": 25,
  "affected_sellers": ["18888", "19999", "20000"],
  "status": "open"
}
```

---

## 🔌 API Endpoints

### Transcript Operations
| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/ingest` | Submit new transcript for analysis |
| `POST` | `/analyze` | Analyze transcript without storing |
| `GET` | `/calls/{id}` | Get analysis for specific call |

### Seller Profiles
| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/sellers` | List all sellers with health status |
| `GET` | `/sellers/{id}` | Get detailed seller profile |

### Analytics
| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/aggregates` | List available aggregate dates |
| `GET` | `/aggregates/{date}` | Get daily aggregate data |
| `POST` | `/aggregate` | Trigger manual aggregation |

### Tickets
| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/tickets` | List ticket dates |
| `GET` | `/tickets/{date}` | Get tickets for specific date |

### Utility
| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/health` | Health check |
| `GET` | `/` | Dashboard UI |

---

## 🖥️ Dashboard UI

### Tab 1: Call Simulator
- Paste transcript JSON or text
- Manually trigger analysis
- View real-time processing logs
- See analysis results immediately

### Tab 2: Seller Profiles
- List all tracked sellers
- Search by seller ID
- View health scores (Critical/At Risk/Healthy)
- Click seller to see full profile
- Click call ID to see call analysis

### Tab 3: Analytics Dashboard
- **Hero Stats**: Total calls, sellers, issues, tickets
- **Issue Categories**: Bar chart of top issues
- **Sentiment Analysis**: Donut chart
- **Churn Risk**: Risk distribution
- **Critical Sellers**: Sellers needing attention
- **Upsell Opportunities**: Revenue potential

### Tab 4: Tickets
- Auto-generated tickets
- Priority-sorted
- Affected seller IDs
- Clickable seller links

---

## ⚙️ Setup & Configuration

### Prerequisites
- Go 1.24+
- Google Gemini API Key
- MongoDB Atlas Account (optional but recommended)

### Environment Variables
```bash
# Required
export GEMINI_API_KEY="your-gemini-api-key"

# Optional (for persistence)
export MONGODB_URI="mongodb+srv://user:pass@cluster.mongodb.net/"

# Optional (for demo mode)
export DEMO_MODE="true"  # Disables watcher, uses existing data
```

### Running the Server
```bash
# Build
go build -o im-ai-voice .

# Run (normal mode - processes new transcripts)
GEMINI_API_KEY="..." MONGODB_URI="..." ./im-ai-voice

# Run (demo mode - uses existing data)
DEMO_MODE=true GEMINI_API_KEY="..." MONGODB_URI="..." ./im-ai-voice
```

### Adding Transcripts
Place JSON files in `data/transcripts/` with format:
```
gluser_{seller_id}_call_{call_id}.json
```

The watcher will automatically detect and process them.

---

## 📝 How It Works - Step by Step

### Step 1: Transcript Arrives
A new call transcript file is placed in `data/transcripts/`

### Step 2: Watcher Detects
The watcher (running every 5 seconds) finds the new file

### Step 3: AI Analysis
The transcript is sent to Google Gemini with a specialized prompt that:
- Understands IndiaMART's business context
- Knows the 17+ feature buckets (Lead Quality, Billing, etc.)
- Extracts structured insights in JSON format

### Step 4: Save Results
- Analysis saved to MongoDB (`call_analyses` collection)
- Seller profile updated (`seller_profiles` collection)
- Health score recalculated

### Step 5: Check Threshold
If 10 new analyses have been completed since last aggregation:
- Run daily aggregation
- Group issues by bucket
- Calculate statistics
- Generate tickets for buckets with 3+ issues

### Step 6: View in Dashboard
Open http://localhost:8080 to see:
- Real-time seller health
- Issue trends
- Auto-generated tickets
- Click on any call to see full analysis

---

## 📈 Feature Buckets

The system categorizes issues into these buckets:

1. **Lead Quality** - Irrelevant/fake leads
2. **Lead Quantity** - Not enough leads
3. **Lead Management** - BuyLead system issues
4. **Billing & Renewal** - Payment problems
5. **Pricing** - Cost concerns
6. **Communication** - Contact issues
7. **Visibility & Ranking** - Search ranking
8. **Catalog & Storefront Setup** - Profile setup
9. **App / Platform Usability** - Technical issues
10. **Promoted Listing & Lead Priority** - TrustSEAL, paid features
11. **Support & Training** - Help requests
12. **Account / Dashboard** - Account issues
13. **Buyer Interaction** - Buyer behavior
14. **Payments** - Transaction issues
15. **Competition** - Competitor mentions
16. **Retention & Churn** - Cancellation requests
17. **Other** - Miscellaneous

---

## 🎯 Business Value

| Metric | Benefit |
|--------|---------|
| **Automated Analysis** | 10-15 seconds per call vs manual review |
| **Issue Detection** | 100% coverage, no missed issues |
| **Churn Prediction** | Proactive retention actions |
| **Upsell Detection** | Revenue opportunities identified |
| **Aggregated Insights** | Pattern recognition across calls |
| **Auto-Ticketing** | Reduced manual ticket creation |

---

## 🔒 Data Flow Summary

```
[Call Recording] → [Transcript] → [Gemini AI] → [Structured Analysis]
                                                        ↓
                                              [Seller Profile Update]
                                                        ↓
                                              [Daily Aggregation]
                                                        ↓
                                              [Auto-Generated Tickets]
                                                        ↓
                                              [Dashboard Visualization]
```

---

## 📞 Support

For questions or issues, contact the development team.

---

*Documentation generated on December 12, 2025*
