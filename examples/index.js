console.log('GitHub Mapper Server Starting...');

const express = require('express');
const app = express();
const port = process.env.PORT || 5050;

app.get('/health', (req, res) => {
  res.json({ status: 'healthy' });
});

app.get('/', (req, res) => {
  res.json({ message: 'GitHub Mapper API' });
});

app.listen(port, () => {
  console.log(`Server running on port ${port}`);
});
