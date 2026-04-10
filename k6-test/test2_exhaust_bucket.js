import http from 'k6/http';
import { sleep } from 'k6';

export const options = {
    vus: 1,
    iterations: 6,
};

export default function () {

    const res = http.get(
        'http://localhost:8080/r/test123',
        {
            redirects: 0
        }
    );

    console.log(
        `Status=${res.status} Remaining=${res.headers['X-Ratelimit-Remaining']}`
    );

    sleep(0.1);
}