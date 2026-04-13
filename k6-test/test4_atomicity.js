import http from 'k6/http';

export const options = {

    vus: 20,
    duration: '10s',
    // use iterations: 10 for better simulation of atomicity

};

export default function () {

    const res = http.get(
        'http://localhost:8080/r/test123',
        {
            redirects: 0
        }
    );

    console.log(res.status);

}