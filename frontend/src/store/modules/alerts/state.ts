import { AlertMessage } from '../../../types/alerts.type';
import { KeyValuePair } from '../../../types/types.type';

export interface AlertModuleState {
  messages: KeyValuePair<AlertMessage>;
}
