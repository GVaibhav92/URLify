import http from "k6/http";
import { sleep, check } from "k6";

const BASE_URL = "http://localhost:8080";
const TOKEN = "";

export const options = {
  stages: [
    { duration: "10s", target: 1 },   // warmup
    { duration: "30s", target: 10 },
    { duration: "1m", target: 30 },
    { duration: "1m", target: 50 },
    { duration: "30s", target: 0 },
  ],
  thresholds: {
    http_req_duration: ["p(95)<100"],
    http_req_failed: ["rate<0.01"],
  },
};

export function setup() {
  const codes = [];

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
      console.error(`Failed to create URL: ${res.status}`);
    }
  }

  console.log(`Created ${codes.length} benchmark URLs`);

  return { codes };
}

export default function (data) {
  const codes = data.codes;

  if (!codes || codes.length === 0) {
    console.error("No short codes available");
    return;
  }

  const code =
    codes[Math.floor(Math.random() * codes.length)];

  const res = http.get(
    `${BASE_URL}/r/${code}`,
    { redirects: 0 }
  );

  check(res, {
    "status is 301": (r) => r.status === 301,
  });

  sleep(0.1);
}