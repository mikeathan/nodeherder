export interface HelpTopic {
  readonly id: string;
  readonly title: string;
  readonly paragraphs: readonly string[];
}

/** Guidance shared by every integration. Capability-specific notes live with their adapter. */
export const COMMON_HELP: readonly HelpTopic[] = [
  {
    id: 'start',
    title: 'Set up an integration',
    paragraphs: [
      'Open Integrations, choose Add integration, select an integration type and provider, then complete its setup. Give the saved integration a name you will recognise later.',
      'A provider may take a moment to accept or finish a lookup. Check status continues the same operation; Refresh retrieves newer data and never changes the saved setup.',
    ],
  },
  {
    id: 'files',
    title: 'Where changes are saved',
    paragraphs: [
      'Configure operations edits the provider, not a saved integration. Form and YAML edit the same host configuration file. Saving validates the file and applies it to every integration using that provider.',
      'Setup choices, provider references, credentials and results are stored privately in the data volume. They are not added to the provider configuration file. You can commit provider files using your usual Git tools; NodeHerder does not commit them for you.',
      'Up to 32 recently visited provider drafts are kept in memory while you are signed in. Save important changes before leaving the application; drafts are cleared on sign out.',
      'Reload from host reads changes made outside the UI. Restore previous writes the previous good configuration as a new revision. Avoid editing the host file and saving in the UI at the same time.',
    ],
  },
  {
    id: 'values',
    title: 'Choose where request values come from',
    paragraphs: [
      'Provider field is the parameter name the external API expects. Use this value chooses the information to send: for example a search term entered in Setup or an option selected in Setup.',
      'A saved provider reference is collected from an earlier response and reused automatically. You do not need to type lookup IDs. A fixed value is a constant included in the shareable file; do not put a token or a personal location there.',
      'URL path segments are the pieces after the provider address. URL parameters are added to the address as a query. Submitted fields form the JSON body of a POST request. The Form controls handle the request encoding.',
    ],
  },
  {
    id: 'steps',
    title: 'How the requests connect',
    paragraphs: [
      'Requests have names you choose. Setup selects which named request performs the initial lookup, which one retrieves the selected result, and which one checks its status. Refresh selects the request that reads updated data.',
      'Values saved from the response declare which provider fields to remember. Save when the response arrives for a lookup reference; use When the lookup finishes successfully for a value that is only returned at completion.',
      'The initial lookup cannot use a value saved by a later request. Validation explains missing or unavailable values before any configuration is saved. New operations, authentication protocols and information types require application support.',
    ],
  },
  {
    id: 'fields',
    title: 'Map the response fields',
    paragraphs: [
      'A field location points to a value in the provider response. /items means a list, /name means a value inside each item, and /result selects a nested section.',
      'Map the list once. NodeHerder applies the same mapping to every returned item, including new items that use the same shape. You do not need a configuration section per item.',
      'These locations use JSON Pointer: a slash separates each level. If the provider moves a field into another object, update its location, for example /details/name. This is a small mapping of the fields we use, not a copy of the provider’s entire schema.',
      'A missing value or an empty list is not proof that something was cancelled or unavailable. Approximate or neighbouring matches are rejected when the provider supplies an exact selection.',
    ],
  },
  {
    id: 'repair',
    title: 'Repair a changed URL or response',
    paragraphs: [
      'For a changed provider address, edit Provider URL. For a changed endpoint, open its request and edit the URL path or parameters. For a renamed response field, update its field location or saved values.',
      'Validate checks the configuration without sending a request. Save changes applies it. Then open an existing integration and choose Refresh to check the provider. A failed refresh keeps the last successful result.',
      'Changing a destination requires review before saved inputs or an existing credential can be sent there. The YAML itself cannot approve a credential transfer. If a new login method or a different response transformation is required, the application may need an update.',
    ],
  },
  {
    id: 'errors',
    title: 'Understand status and errors',
    paragraphs: [
      'Pending means the provider has accepted the lookup but has not finished it. Check status continues it. A request limit includes the next time you can try; repeated clicks do not bypass it.',
      'An expired setup must be started again. A changed-in-another-session message means another user or host edit saved a newer revision; reload and review before saving.',
      'An old schedule stays visible after a failed refresh. Last checked is when NodeHerder checked the provider; provider verification is the source’s own timestamp. A successful unchanged response does not make old source data fresh.',
      'If Integrations is unavailable, check the persistent config/data mounts and separately mounted private keyring. Never replace a lost key with a new one and expect old data to decrypt; restore the matching key backup.',
    ],
  },
] as const;

/** Case-insensitive search across a help topic's title and body. */
export function matchesHelpTopic(topic: HelpTopic, search: string): boolean {
  const needle = search.trim().toLowerCase();
  return !needle || `${topic.title} ${topic.paragraphs.join(' ')}`.toLowerCase().includes(needle);
}
