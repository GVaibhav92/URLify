import http from "k6/http";
import { sleep, check } from "k6";

const BASE_URL = "http://localhost:8080";  
const TOKEN = "";  

export const options = {
  // Load stage profile: 50 concurrent VUs
  stages: [
    { duration: "10s", target: 1 },   //warmup
    { duration: "30s", target: 10 },  //ramp
    { duration: "1m", target: 30 },   //sustain 
    { duration: "1m", target: 50 },   //peak
    { duration: "30s", target: 0 },   //shutdown
  ],
  thresholds: {
    http_req_duration: ["p(95)<100"],  //95th percentile response time < 100ms
    http_req_failed: ["rate<0.01"],    //error rate must be < 1%
  },
};

//Pre-create test data
export function setup() {
  const codes = [];  

  // Create 30 unique short URLs for testing redirects
  for (let i = 0; i < 30; i++) {
    const res = http.post(
      `${BASE_URL}/urls`,  
      JSON.stringify({
        original_url: `https://google.com/${i}`, 
      }),
      {
        headers: {
          Authorization: `Bearer ${TOKEN}`,    
          "Content-Type": "application/json",
        },
      }
    );
    if (res.status === 201) {
      const body = JSON.parse(res.body);
      codes.push(body.short_code);
    } else {
      console.error(`Failed to create URL ${i}: ${res.status}`);
    }
  }

  console.log(`Created ${codes.length} benchmark URLs`); 
  return { codes }; 
}

//Each VU repeatedly tests random short URL redirects
export default function (data) {
  const codes = data.codes;  

  if (!codes || codes.length === 0) {
    console.error("No short codes available - setup failed");
    return;
  }

  // Pick random short code for this iteration
  const code = codes[Math.floor(Math.random() * codes.length)];

  // Test redirect endpoint: GET /r/{short_code}
  const res = http.get(
    `${BASE_URL}/r/${code}`,
    { redirects: 0 } 
  );

  // Verify expected 301 Moved Permanently redirect status
  check(res, {
    "status is 301": (r) => r.status === 301,
  });

  sleep(0.1);  
}