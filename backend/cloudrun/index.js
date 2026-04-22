const { spawn } = require('child_process');
const path = require('path');

exports.main = async (event, context) => {
  const scriptPath = path.join(__dirname, 'sqlite-manager');
  
  return new Promise((resolve, reject) => {
    const args = process.argv.slice(2);
    const child = spawn(scriptPath, args, {
      stdio: ['pipe', 'pipe', 'pipe']
    });

    let stdout = '';
    let stderr = '';

    child.stdout.on('data', (data) => {
      stdout += data.toString();
    });

    child.stderr.on('data', (data) => {
      stderr += data.toString();
    });

    child.on('close', (code) => {
      resolve({
        statusCode: 200,
        body: JSON.stringify({ stdout, stderr, code })
      });
    });

    child.on('error', (err) => {
      reject(err);
    });

    // 如果是 HTTP 请求，代理到标准输入
    if (event.httpMethod) {
      const requestBody = event.body || '';
      child.stdin.write(requestBody);
      child.stdin.end();
    }
  });
};
