import { sleep } from 'k6';
import { visit } from '../lib/requests.js';

// Runs for 3s so that k6 is still running when the e2e test looks for its process.
// See runK6 in e2e/integration_test.go for why a script without a duration is too short.
export const options = { vus: 1, duration: '3s' };

export default function () {
	visit('https://www.steadybit.com');
	sleep(1);
}
