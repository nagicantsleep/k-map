# BRD: OSM-Backed Geolocation SaaS

## 1. Document Control

- Product: `k-map`
- Version: `0.1`
- Status: Draft
- Date: `2026-03-09`

## 2. Executive Summary

`k-map` is a SaaS geolocation service built on OpenStreetMap data to serve as a commercial alternative to Google Maps geocoding for backend and product teams that need:

1. Reverse geocoding: latitude/longitude to address or place
2. Forward geocoding: address or place to latitude/longitude
3. Proximity validation: determine whether a coordinate is near a target address or place

The business goal is to provide a lower-cost, controllable, privacy-conscious location API with predictable pricing and minimal vendor lock-in.

## 3. Problem Statement

Teams commonly use Google Maps because it is easy to start with, but they face recurring issues:

- High and sometimes unpredictable cost at scale
- Vendor lock-in
- Limited control over data freshness, ranking, and operational behavior
- Compliance or privacy concerns for customer location data
- Difficulty tailoring proximity rules to business logic

There is demand for a simpler, backend-first geolocation service based on open map data with commercial ownership of the service layer.

## 4. Business Opportunity

An OSM-backed geolocation service can compete on:

- Lower total cost for medium to high volume API traffic
- Transparent behavior and configurable ranking logic
- Self-hosted or dedicated deployment options for regulated customers
- Narrow, reliable feature set instead of a large mapping platform
- Faster integration for SaaS products that only need geocoding and validation

## 5. Goals

### Primary Goals

- Launch an MVP that exposes reliable forward geocoding, reverse geocoding, and proximity validation APIs
- Achieve production-readiness for multi-tenant SaaS usage
- Keep initial architecture minimal and operable by a small team

### Secondary Goals

- Support future extensibility for autocomplete, batch jobs, and regional ranking improvements
- Preserve the option to self-host core geocoding data infrastructure

## 6. Non-Goals

- Turn-by-turn navigation
- Full map tile rendering platform for end users
- Route optimization
- Places reviews, photos, or rich POI content
- Consumer map UI as part of MVP
- Global perfect parity with Google Maps result quality

## 7. Target Customers

### Primary Segments

- B2B SaaS products that capture customer addresses
- Logistics or delivery platforms validating pickup and dropoff coordinates
- CRM, field service, and operations software
- Platforms migrating away from Google Maps due to cost or policy constraints

### Buyer Personas

- Engineering manager reducing infra/API cost
- Product manager needing address workflows
- Platform team needing a stable internal geolocation API

## 8. User Needs

Customers need to:

- Convert user-entered addresses into normalized coordinates
- Convert device or event coordinates into a usable address
- Check whether a reported coordinate is close enough to an expected location
- Rely on a documented API with stable semantics and quota controls
- Avoid depending on public community-operated Nominatim endpoints for commercial traffic

## 9. Value Proposition

`k-map` offers:

- OSM-backed geocoding APIs built for SaaS workloads
- Predictable API contract and account-level controls
- Lower platform cost than premium proprietary mapping stacks
- Configurable proximity rules for operational workflows
- Ability to evolve from shared SaaS to dedicated deployments

## 10. Business Requirements

### Functional Requirements

- Expose API endpoints for forward geocoding, reverse geocoding, and proximity checks
- Support structured and unstructured address input
- Return normalized, ranked results with confidence metadata
- Support tenant authentication and quota enforcement
- Log requests and outcomes for billing, support, and observability

### Non-Functional Requirements

- Production uptime target suitable for B2B API usage
- Low operational complexity for MVP
- Horizontal scaling at the stateless API layer
- Clear data refresh process for OSM imports
- Privacy-conscious handling of location data

## 11. Success Metrics

### Business KPIs

- Time to first successful API integration
- Monthly active API tenants
- Cost per 1,000 requests
- Gross margin by traffic tier
- Trial-to-paid conversion

### Product KPIs

- Reverse geocode success rate
- Forward geocode match rate on target datasets
- Proximity decision accuracy against labeled test cases
- P95 API latency by endpoint
- Error rate by endpoint

## 12. Constraints and Assumptions

### Constraints

- Service must be minimal in initial scope
- Product is greenfield, so team should avoid over-engineering
- OSM-derived data quality varies by country and locality
- Public Nominatim infrastructure is not suitable for commercial SaaS dependence

### Assumptions

- Initial customers can tolerate address quality that is strong but not Google-level in every market
- MVP can prioritize one or a few regions before full global optimization
- Customers primarily need backend APIs, not consumer map visualization

## 13. Recommended Business Approach

For MVP, the company should operate a managed API service backed by self-hosted OSM geocoding infrastructure. This avoids dependence on third-party public endpoints and creates a direct path to commercial control, pricing, and reliability.

Recommended packaging:

- Free trial with strict quota
- Usage-based paid plans
- Enterprise dedicated environment later

## 14. Major Risks

- OSM data coverage and ranking quality may underperform in some markets
- Data import and indexing can be operationally heavy
- Proximity validation can be misunderstood if semantics are not explicit
- Customers may compare results directly to Google and expect parity

## 15. Risk Mitigations

- Define clear regional support tiers
- Publish confidence and match metadata
- Make proximity logic explicit and configurable
- Start with a minimal, well-tested API surface
- Maintain a repeatable import and rollback process for OSM refreshes

## 16. High-Level Delivery Phases

### Phase 1: MVP

- Forward geocoding API
- Reverse geocoding API
- Proximity validation API
- API keys, quotas, logging, and basic billing hooks

### Phase 2

- Batch geocoding
- Better ranking and region tuning
- Address parsing improvements
- Customer dashboard

### Phase 3

- Dedicated regions or customer deployments
- SLA-backed enterprise tier
- Advanced match explanation and audit tooling

## 17. Approval Criteria

The BRD is satisfied when the MVP:

- Solves the three core location use cases reliably
- Is commercially operable as a multi-tenant API service
- Has a clear path to profitable scaling
- Avoids unnecessary platform scope outside geolocation fundamentals
