import http from 'k6/http';
import { check, sleep } from 'k6';

const BASE_URL = 'http://localhost:8080';

const TOKEN = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMDNlZDU0NGYtOTgyYy00ZGIwLWJjNWEtODM2MTRmNGNmNmQxIiwicm9sZSI6ImZyZWUiLCJleHAiOjE3NzYyMzMzNTIsImlhdCI6MTc3NjE0Njk1Mn0.9zSoP8Cw7arsXBZONQ9gc3gry5xN9xbAjsxWXdYhafA';

const SHORT_CODES = [];

export const options = {
  stages: [
    { duration: '10s', target: 1 },   // warmup
    { duration: '30s', target: 10 },   // ramp up
    { duration: '1m',  target: 30 },   // sustained load
    { duration: '30s', target: 50 },   // spike
    { duration: '30s', target: 0  },   // ramp down
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'],   // 95% of requests under 500ms
    http_req_failed:   ['rate<0.995'],   
  },
};

//Setup — runs once, creates test URLs 
export function setup() {
  const codes = [];

  for (let i = 0; i < 5; i++) {
    const res = http.post(
      `${BASE_URL}/urls`,
      JSON.stringify({ original_url: 'https://www.google.com' }),
      {
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${TOKEN}`,
        },
      }
    );

    if (res.status === 201) {
      const body = JSON.parse(res.body);
      codes.push(body.short_code);
    }
  }

  console.log(`Created ${codes.length} test URLs: ${codes.join(', ')}`);
  return { codes };
}

//Main VU loop
export default function (data) {
  const codes = data.codes;

  if (codes.length === 0) {
    console.error('No short codes available');
    return;
  }

  // Pick a random short code
  const code = codes[Math.floor(Math.random() * codes.length)];

  // 70% redirects — main traffic pattern
  if (Math.random() < 0.70) {
    const res = http.get(`${BASE_URL}/r/${code}`, {
      redirects: 0, // don't follow — just test our endpoint
    });

    check(res, {
      'redirect status 301': (r) => r.status === 301,
      'has Location header': (r) => r.headers['Location'] !== undefined,
    });
  }
  // 15% list URLs (authenticated)
  else if (Math.random() < 0.85) {
    const res = http.get(`${BASE_URL}/urls`, {
      headers: { 'Authorization': `Bearer ${TOKEN}` },
    });

    check(res, {
      'list status 200': (r) => r.status === 200,
    });
  }
  // 15% stats endpoint
  else {
    const res = http.get(`${BASE_URL}/stats`);
    check(res, {
      'stats status 200': (r) => r.status === 200,
    });
  }

  sleep(0.2); // 100ms between requests per VU
}

//Teardown-runs once after test 
export function teardown(data) {
  console.log(`Load test complete. Tested with codes: ${data.codes.join(', ')}`);
}