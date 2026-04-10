import http from 'k6/http';
import { check } from 'k6';

export const options = {
    vus: 1,
    iterations: 1,
};

export default function () {

    const res = http.get(
        'http://localhost:8080/r/test123',
        {
            redirects: 0
        }
    );

    console.log("Status:", res.status);
    console.log("Limit:", res.headers['X-Ratelimit-Limit']);
    console.log("Remaining:", res.headers['X-Ratelimit-Remaining']);
    console.log("Refill:", res.headers['X-Ratelimit-Refill-Rate']);

    check(res, {
        'status returned': (r) => r.status !== 0,
        'limit header exists': (r) =>
            r.headers['X-Ratelimit-Limit'] !== undefined,
        'remaining header exists': (r) =>
            r.headers['X-Ratelimit-Remaining'] !== undefined,
    });
}