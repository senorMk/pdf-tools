import frontendPackage from '../package.json';
import desktopConfig from '../../wails.json';

export const appInfo = {
  name: 'PDF Tools',
  version: frontendPackage.version,
  developer: desktopConfig.author.name,
  license: 'MIT License',
};
