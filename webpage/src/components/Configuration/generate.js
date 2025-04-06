/* This script reads a help.txt file and generates a help.json file with the configuration options. */
const fs = require('fs');

const parseConfigOptions = (text) => {
  const lines = text.split('\n');
  return lines.map(line => {
    const match = line.match(/--([\w.]+)(?:\s(\w+))?\s+(.*?)(?:\s+\(default\s+"?([^"]+)"?\))?$/);
    if (match) {
      let [, key, type, description, defaultValue] = match;
      // if the type is not specified, it is a boolean
      if (!type) {
        type = 'boolean';
        // if the default value is not specified, it is false
        if (!defaultValue) {
          defaultValue = 'false';
        } else if (defaultValue !== 'false') {
          defaultValue = 'true';
        }
      }
      // this is an opaque object
      if (type === 'string' && defaultValue === '{}') {
        type = 'object';
        defaultValue = {};
      }
      // this is an opaque array
      if (type === 'string' && defaultValue === '[]') {
        type = 'array';
        defaultValue = [];
      }
      return { key: key.split('.'), type, defaultValue: defaultValue || undefined, description };
    }
    return null;
  }).filter(option => option !== null);
};

const [,, inputFilePath, outputFilePath] = process.argv;

if (!inputFilePath || !outputFilePath) {
  console.error('Usage: node generate.js <inputFilePath> <outputFilePath>');
  process.exit(1);
}

fs.readFile(inputFilePath, 'utf8', (err, data) => {
  if (err) {
    console.error('Error reading input file:', err);
    return;
  }
  const configOptions = parseConfigOptions(data);
  fs.writeFile(outputFilePath, JSON.stringify(configOptions, null, 2), (err) => {
    if (err) {
      console.error('Error writing output file:', err);
    } else {
      console.log('Output file generated successfully.');
    }
  });
});
