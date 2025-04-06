import React from 'react';
import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';
import CodeBlock from '@theme/CodeBlock';

interface ConfigOptionValue {
  type?: string;
  description?: string;
  defaultValue?: string;
}

interface ConfigOption extends ConfigOptionValue {
  key: string[];
}

function configKey(key: string | string[], value: ConfigOptionValue | any): ConfigOption {
  if (typeof key === 'string') {
    key = key.split('.');
  }
  if (typeof value === 'object') {
    return {
      key,
      type: value.type || getType(value.defaultValue),
      description: value.description,
      defaultValue: value.defaultValue,
    }
  } else {
    return {
      key,
      type: getType(value),
      defaultValue: value,
    }
  }
}

function configKeys(optValues: Record<string, ConfigOptionValue | any>): ConfigOption[] {
  let options: ConfigOption[] = [];
  Object.entries(optValues).forEach(([key, value]) => {
    options.push(configKey(key, value));
  });
  return options;
}

function filterKeys(options: ConfigOption[], filter: string): ConfigOption[] {
  return options.filter(option => {
    const key = option.key.join('.');
    return key.startsWith(filter)
  });
}

function defaultValue(value: ConfigOptionValue): string {
  switch (value.type) {
    case 'boolean':
      return `${value.defaultValue || false}`;
    case 'int':
    case 'float':
    case 'number':
      return `${value.defaultValue || 0}`;
    case 'duration':
    case 'string':
      return `${value.defaultValue ? `"${value.defaultValue}"` : '<string>'}`;
    case 'strings':
      return '<comma-separated list of strings>';
    case 'object':
      return '<json encoded object>';
    case 'array':
      return '<json encoded array>';
    default:
      return value.type ? `<${value.type}>` : '';
  }
}

function getType(value: any): string {
  if (Array.isArray(value)) {
    return 'array';
  }
  return typeof value;
}

export function EnvironmentVariables({ options, comments, ...props }: { options: ConfigOption[], comments?: boolean }) {
  if (typeof comments === 'undefined') {
    comments = true;
  }

  let code = '';
  options.forEach(option => {
    const description = option.description ? option.description : '';
    const type = option.type ? ` (${option.type})` : '';
    if (comments && description) {
      code += `# ${description}${type}\n`;
    }
    code += `NEKO_${option.key.join('_').toUpperCase()}=${defaultValue(option)}\n`;
  });

  return (
    <CodeBlock language="shell" {...props}>
      {code}
    </CodeBlock>
  );
}

export function CommandLineArguments({ options, comments, ...props }: { options: ConfigOption[], comments?: boolean }) {
  if (typeof comments === 'undefined') {
    comments = true;
  }

  let code = '';
  options.forEach(option => {
    const description = option.description ? option.description : '';
    const type = option.type ? ` (${option.type})` : '';
    if (comments && description) {
      code += `# ${description}${type}\n`;
    }
    code += `--${option.key.join('.')} ${defaultValue(option)}\n`;
  });

  return (
    <CodeBlock language="shell" {...props}>
      {code}
    </CodeBlock>
  );
}

export function YamlFileContent({ options, comments, ...props }: { options: ConfigOption[], comments?: boolean }) {
  if (typeof comments === 'undefined') {
    comments = true;
  }

  const final = Symbol('final');

  const buildYaml = (obj: Record<string, any>, prefix = '') => {
    let code = '';
    Object.entries(obj).forEach(([key, option]) => {
      if (typeof option === 'object' && !Array.isArray(option) && !option[final]) {
        code += prefix+`${key}:\n`;
        code += buildYaml(option, prefix + '  ');
      } else {
        const description = option.description ? option.description : '';
        const type = option.type ? ` (${option.type})` : '';
        if (comments && description) {
          code += `${prefix}# ${description}${type}\n`;
        }
        let value: string;
        if (option.type === 'strings') {
          value = option.defaultValue ? `[ "${option.defaultValue}" ]` : '[ <string> ]';
        } else if (option.type === 'object') {
          value = "{}"
        } else if (option.type === 'array') {
          value = "[]"
        } else {
          value = defaultValue(option);
        }
        code += `${prefix}${key}: ${value}\n`;
      }
    });
    return code;
  };

  const yamlCode = buildYaml(options.reduce((acc, option) => {
    const keys = option.key;
    let current = acc;
    keys.forEach((key, index) => {
      if (!current[key]) {
        current[key] = index === keys.length - 1 ? option : {};
      }
      current = current[key];
    });
    current[final] = true;
    return acc;
  }, {}));

  return (
    <CodeBlock language="yaml" {...props}>
      {yamlCode}
    </CodeBlock>
  );
}

type ConfigurationTabProps = {
  options?: ConfigOption[] | Record<string, ConfigOptionValue | any>;
  heading?: boolean;
  comments?: boolean;
  filter?: string | string[] | Record<string, ConfigOptionValue | any>;
};

export function ConfigurationTab({ options, heading, comments, filter, ...props }: ConfigurationTabProps) {
  var configOptions: ConfigOption[] = [];
  if (Array.isArray(options)) {
    configOptions = options;
  } else {
    configOptions = configKeys(options)
  }
  if (typeof comments === 'undefined') {
    comments = true;
  }
  if (typeof heading === 'undefined') {
    heading = false;
  }

  if (Array.isArray(filter)) {
    let filteredOptions: ConfigOption[] = [];
    for (const f of filter) {
      filteredOptions = [ ...filteredOptions, ...filterKeys(configOptions, f) ];
    }
    configOptions = filteredOptions;
  } else if (typeof filter === 'string') {
    configOptions = filterKeys(configOptions, filter);
  } else if (typeof filter === 'object') {
    let filteredOptions: ConfigOption[] = [];
    for (const k in filter) {
      let filtered = configOptions.find(option => {
        return option.key.join('.') === k;
      });
      let replaced = configKey(k, filter[k]);
      filteredOptions = [ ...filteredOptions, { ...filtered, ...replaced } ];
    }
    configOptions = filteredOptions;
  }

  return (
    <Tabs groupId="configuration" defaultValue="yaml" values={[
      { label: 'YAML Configuration File', value: 'yaml' },
      { label: 'Environment Variables', value: 'env' },
      { label: 'Command Line Arguments', value: 'args' },
    ]} {...props}>
      <TabItem value="env" label="Environment Variables">
        {heading && (
          <p>You can set the following environment variables in your <code>docker-compose.yaml</code> file or in your shell environment.</p>
        )}
        {EnvironmentVariables({ options: configOptions, comments })}
      </TabItem>
      <TabItem value="args" label="Command Line Arguments">
        {heading && (
          <p>You can list the following command line arguments using <code>neko serve --help</code>.</p>
        )}
        {CommandLineArguments({ options: configOptions, comments })}
      </TabItem>
      <TabItem value="yaml" label="YAML Configuration File">
        {heading && (
          <p>You can create a <code>/etc/neko/neko.yaml</code> file with the following configuration options.</p>
        )}
        {YamlFileContent({ options: configOptions, comments })}
      </TabItem>
    </Tabs>
  );
}
