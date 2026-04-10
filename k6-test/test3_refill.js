import http from 'k6/http';
import { sleep } from 'k6';

export default function () {

    console.log("Exhausting bucket...");

    for (let i = 0; i < 5; i++) {

        http.get(
            'http://localhost:8080/r/test123',
            {
                redirects: 0
            }
        );

    }

    console.log("Waiting 3 seconds...");

    sleep(3);

    const res = http.get(
        'http://localhost:8080/r/test123',
        {
            redirects: 0
        }
    );

    console.log(
        `After refill → Status=${res.status} Remaining=${res.headers['X-Ratelimit-Remaining']}`
    );
}