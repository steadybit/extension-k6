import http from 'k6/http';
import { Counter } from 'k6/metrics';

// Only recorded when this module, imported via a relative path, ran from the archive. The e2e
// test looks for it in the metrics k6 writes.
const moduleCalls = new Counter('archive_module_calls');

export function visit(url) {
	moduleCalls.add(1);
	return http.get(url);
}
