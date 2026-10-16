// A script that defines its own k6 scenario. Run as is, k6 runs exactly this scenario. Combined with the step's
// Duration or Virtual users parameter, the extension refuses to run it (see the e2e tests).
import http from 'k6/http';
import { sleep } from 'k6';

export const options = {
  scenarios: {
    smoke: {
      executor: 'constant-vus',
      vus: 1,
      duration: '3s',
    },
  },
};

export default function smoke() {
  http.get('https://www.steadybit.com');
  sleep(1);
}
