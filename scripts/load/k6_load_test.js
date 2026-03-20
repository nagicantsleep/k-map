/**
 * k-map load test — all three geocoding endpoints
 *
 * Usage:
 *   k6 run scripts/load/k6_load_test.js
 *
 * Environment variables:
 *   KMAP_API_URL   Base URL of the API (default: http://localhost:8080)
 *   KMAP_API_KEY   API key to use for authenticated requests (default: test-key)
 *
 * PRD latency targets:
 *   Forward geocode  P95 < 800 ms
 *   Reverse geocode  P95 < 500 ms
 *   Proximity check  P95 < 1000 ms
 */

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Counter, Trend } from 'k6/metrics';

// --- Configuration ---

const BASE_URL = __ENV.KMAP_API_URL || 'http://localhost:8080';
const API_KEY  = __ENV.KMAP_API_KEY  || 'test-key';

const headers = {
  'Content-Type': 'application/json',
  'X-API-Key': API_KEY,
};

// --- Custom metrics ---

const forwardErrors   = new Counter('forward_errors');
const reverseErrors   = new Counter('reverse_errors');
const proximityErrors = new Counter('proximity_errors');

const forwardDuration   = new Trend('forward_duration',   true);
const reverseDuration   = new Trend('reverse_duration',   true);
const proximityDuration = new Trend('proximity_duration', true);

// --- Load test options ---

export const options = {
  scenarios: {
    load: {
      executor: 'constant-vus',
      vus: 50,
      duration: '2m',
    },
  },
  thresholds: {
    // PRD P95 latency targets
    'forward_duration{p(95)}':   ['p(95)<800'],
    'reverse_duration{p(95)}':   ['p(95)<500'],
    'proximity_duration{p(95)}': ['p(95)<1000'],

    // 5xx error rate must stay under 1%
    'http_req_failed': ['rate<0.01'],

    // Custom error counters should stay at 0 for a clean run
    'forward_errors':   ['count<5'],
    'reverse_errors':   ['count<5'],
    'proximity_errors': ['count<5'],
  },
};

// --- Fixture data for Monaco ---

const forwardQueries = [
  'Mairie de Monaco',
  'Place du Casino, Monaco',
  'Palais Princier de Monaco',
  'Stade Louis II Monaco',
  'Jardin Exotique de Monaco',
];

const reverseCoordinates = [
  { latitude: 43.7311, longitude: 7.4197 },
  { latitude: 43.7393, longitude: 7.4271 },
  { latitude: 43.7353, longitude: 7.4247 },
  { latitude: 43.7272, longitude: 7.4133 },
  { latitude: 43.7338, longitude: 7.4196 },
];

const proximityRequests = [
  { latitude: 43.7311, longitude: 7.4197, target_query: 'Mairie de Monaco',        threshold_meters: 500 },
  { latitude: 43.7393, longitude: 7.4271, target_query: 'Place du Casino, Monaco', threshold_meters: 200 },
  { latitude: 43.7311, longitude: 7.4197, target_query: 'Place du Casino, Monaco', threshold_meters: 100 },
  { latitude: 43.7353, longitude: 7.4247, target_query: 'Palais Princier de Monaco', threshold_meters: 50 },
  { latitude: 43.7272, longitude: 7.4133, target_query: 'La Condamine Monaco',     threshold_meters: 2000 },
];

// --- Virtual user iteration ---

export default function () {
  const idx = Math.floor(Math.random() * 5);

  // --- Forward geocode ---
  const fwBody = JSON.stringify({ query: forwardQueries[idx], limit: 3 });
  const fwRes  = http.post(`${BASE_URL}/v1/geocode/forward`, fwBody, { headers });
  forwardDuration.add(fwRes.timings.duration);

  const fwOk = check(fwRes, {
    'forward: status 200': (r) => r.status === 200,
    'forward: has results field': (r) => {
      try { return Array.isArray(JSON.parse(r.body).results); } catch { return false; }
    },
  });
  if (!fwOk) forwardErrors.add(1);

  sleep(0.1);

  // --- Reverse geocode ---
  const coords = reverseCoordinates[idx];
  const rvBody = JSON.stringify(coords);
  const rvRes  = http.post(`${BASE_URL}/v1/geocode/reverse`, rvBody, { headers });
  reverseDuration.add(rvRes.timings.duration);

  const rvOk = check(rvRes, {
    'reverse: status 200': (r) => r.status === 200,
    'reverse: has latitude field': (r) => {
      try { return typeof JSON.parse(r.body).latitude === 'number'; } catch { return false; }
    },
  });
  if (!rvOk) reverseErrors.add(1);

  sleep(0.1);

  // --- Proximity check ---
  const px    = proximityRequests[idx];
  const pxBody = JSON.stringify(px);
  const pxRes  = http.post(`${BASE_URL}/v1/geocode/proximity`, pxBody, { headers });
  proximityDuration.add(pxRes.timings.duration);

  const pxOk = check(pxRes, {
    'proximity: status 200': (r) => r.status === 200,
    'proximity: has is_near field': (r) => {
      try { return typeof JSON.parse(r.body).is_near === 'boolean'; } catch { return false; }
    },
  });
  if (!pxOk) proximityErrors.add(1);

  sleep(0.2);
}
