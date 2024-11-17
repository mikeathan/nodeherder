import { AlertMessage } from '../../../types/alerts.type';
import { KeyValuePair } from '../../../types/types';

export interface AlertModuleState {
  messages: KeyValuePair<AlertMessage>;
}
