// Visionary RAG Pipeline - k6 Load Test
// Tests 1,000 concurrent users with TTFT < 500ms SLA

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';

// Custom metrics
const ttftMetric = new Trend('ttft_ms');
const errorRate = new Rate('errors');
const successRate = new Rate('success');

// Test configuration
export const options = {
  // Gradual ramp-up to 1,000 concurrent users
  stages: [
    { duration: '1m', target: 100 },   // Ramp to 100 users
    { duration: '2m', target: 500 },   // Ramp to 500 users
    { duration: '3m', target: 1000 },  // Ramp to 1000 users
    { duration: '5m', target: 1000 },  // Stay at 1000 users
    { duration: '2m', target: 500 },   // Ramp down to 500
    { duration: '2m', target: 0 },     // Ramp down to 0
  ],
  
  // Performance thresholds
  thresholds: {
    'http_req_duration': ['p(95)<500'],  // 95% of requests < 500ms
    'ttft_ms': ['p(99)<500'],            // TTFT SLA: 99% < 500ms
    'errors': ['rate<0.01'],             // Error rate < 1%
    'success': ['rate>0.99'],            // Success rate > 99%
  },
  
  // Scenarios
  scenarios: {
    query_test: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '1m', target: 100 },
        { duration: '2m', target: 500 },
        { duration: '3m', target: 1000 },
        { duration: '5m', target: 1000 },
        { duration: '2m', target: 0 },
      ],
      exec: 'queryTest',
    },
  },
};

// Test queries (CBSE Class 8 Science)
const testQueries = [
  { query: 'What is photosynthesis?', grade: 8, subject: 'Science' },
  { query: 'What are the parts of a cell?', grade: 8, subject: 'Science' },
  { query: 'What is force and pressure?', grade: 8, subject: 'Science' },
  { query: 'What is combustion?', grade: 8, subject: 'Science' },
  { query: 'What are microorganisms?', grade: 8, subject: 'Science' },
  { query: 'Explain the structure of an atom', grade: 8, subject: 'Science' },
  { query: 'What is the difference between mitosis and meiosis?', grade: 8, subject: 'Science' },
  { query: 'How does digestion work in humans?', grade: 8, subject: 'Science' },
];

// Generate JWT token (in production, use real JWT)
function generateJWT(grade, subject) {
  return `Bearer mock_jwt_token_grade_${grade}_${subject}`;
}

// Query test function
export function queryTest() {
  // Select random query
  const testData = testQueries[Math.floor(Math.random() * testQueries.length)];
  
  // Generate session ID
  const sessionId = `session_${__VU}_${__ITER}`;
  
  // Prepare request
  const payload = JSON.stringify({
    query: testData.query,
    session_id: sessionId,
  });
  
  const params = {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': generateJWT(testData.grade, testData.subject),
    },
  };
  
  // Make request
  const startTime = Date.now();
  const response = http.post('http://localhost:8080/query', payload, params);
  const endTime = Date.now();
  
  // Calculate metrics
  const totalTime = endTime - startTime;
  const ttft = response.timings ? response.timings.firstByte : totalTime;
  
  // Record metrics
  ttftMetric.add(ttft);
  
  // Check response
  const success = check(response, {
    'status is 200': (r) => r.status === 200,
    'has answer': (r) => r.json('answer') !== '',
    'has sources': (r) => r.json('sources').length > 0,
    'ttft < 500ms': (r) => ttft < 500,
    'total time < 2000ms': (r) => totalTime < 2000,
  });
  
  // Record success/error
  errorRate.add(success ? 0 : 1);
  successRate.add(success ? 1 : 0);
  
  // Small delay between requests
  sleep(0.5);
}

// Health check test
export function healthTest() {
  const response = http.get('http://localhost:8080/health');
  
  check(response, {
    'health status is 200': (r) => r.status === 200,
  });
  
  sleep(1);
}

// Handle summary
export function handleSummary(data) {
  return {
    'data/load_test_results.json': JSON.stringify(data, null, 2),
    stdout: textSummary(data, { indent: ' ', enableColors: true }),
  };
}

// Text summary formatter
function textSummary(data, options) {
  const { metrics } = data;
  
  let summary = '\n========================================\n';
  summary += 'LOAD TEST RESULTS\n';
  summary += '========================================\n\n';
  
  summary += `Total Requests: ${metrics.http_reqs.values.count}\n`;
  summary += `Success Rate: ${(metrics.success.values.rate * 100).toFixed(2)}%\n`;
  summary += `Error Rate: ${(metrics.errors.values.rate * 100).toFixed(2)}%\n\n`;
  
  summary += 'Latency Metrics:\n';
  summary += `  Average: ${metrics.http_req_duration.values.avg.toFixed(2)}ms\n`;
  summary += `  P95: ${metrics.http_req_duration.values['p(95)'].toFixed(2)}ms\n`;
  summary += `  P99: ${metrics.http_req_duration.values['p(99)'].toFixed(2)}ms\n\n`;
  
  summary += 'TTFT Metrics:\n';
  summary += `  Average: ${metrics.ttft_ms.values.avg.toFixed(2)}ms\n`;
  summary += `  P95: ${metrics.ttft_ms.values['p(95)'].toFixed(2)}ms\n`;
  summary += `  P99: ${metrics.ttft_ms.values['p(99)'].toFixed(2)}ms\n\n`;
  
  summary += '========================================\n';
  
  // Check SLA compliance
  const ttftP99 = metrics.ttft_ms.values['p(99)'];
  const errorRate = metrics.errors.values.rate;
  
  if (ttftP99 < 500 && errorRate < 0.01) {
    summary += '✅ SLA MET: TTFT < 500ms, Error Rate < 1%\n';
  } else {
    summary += '❌ SLA NOT MET\n';
    if (ttftP99 >= 500) summary += `   - TTFT P99 (${ttftP99.toFixed(2)}ms) >= 500ms\n`;
    if (errorRate >= 0.01) summary += `   - Error Rate (${(errorRate * 100).toFixed(2)}%) >= 1%\n`;
  }
  
  summary += '========================================\n';
  
  return summary;
}
