import { sleep } from 'k6';
import { visit } from '../lib/requests.js';

// Runs well past the 5s window in which the e2e test looks for the k6 process.
export const options = { vus: 1, duration: '15s' };

export default function () {
	visit('https://www.steadybit.com');
	sleep(1);
}
