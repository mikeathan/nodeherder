export const CONNECTION_HELP = [
  {
    id: 'start',
    title: 'Add bin collections',
    paragraphs: [
      'Open Integrations, choose a provider under Bin collections, then select Add integration. Enter your postcode, choose the exact address and save.',
      'The provider may take a little while. Check status continues the same lookup. Refresh asks for updated dates; this first release does not refresh on a schedule.',
    ],
  },
  {
    id: 'files',
    title: 'Where changes are saved',
    paragraphs: [
      'Configure operations describes the provider, not your address. Form and YAML edit the same host file under configs/integrations/waste_collection/. Save changes validates the file and applies it to every integration using that provider.',
      'Address choices, provider references, tokens and results are stored privately in the data volume. They are not added to the provider YAML. You can commit provider files using your usual Git tools; NodeHerder does not commit them for you.',
      'Up to 32 recently visited provider drafts are kept in memory while you are signed in. Save important changes before leaving the application; drafts are cleared on sign out.',
      'Reload from host reads changes you made outside the UI. Restore previous writes the previous good configuration as a new revision. Avoid editing the host file and saving in the UI at the same time.',
    ],
  },
  {
    id: 'values',
    title: 'Choose where request values come from',
    paragraphs: [
      'Provider field is the parameter name the external API expects. Use this value chooses the information to send: for example Postcode entered in Setup or Address selected in Setup.',
      'A saved provider reference is collected from an earlier response and reused automatically. You do not need to type lookup IDs. A fixed value is a constant included in the shareable file; do not put a token or your address there.',
      'URL path segments are the pieces after the provider address. URL parameters are added to the address as a query. Submitted fields form the JSON body of a POST request. The Form controls handle the request encoding.',
    ],
  },
  {
    id: 'steps',
    title: 'How the requests connect',
    paragraphs: [
      'Requests have names you choose. The setup steps select which named request finds addresses, retrieves the selected address and checks the lookup status. Refresh selects the request that reads a saved schedule.',
      'Values saved from the response declare which provider fields to remember. Save when the response arrives for a lookup reference; use When the lookup finishes successfully for a property token that is only returned at completion.',
      'The initial address search cannot use a value saved by a later request. Validation explains missing or unavailable values before any configuration is saved. New operations, authentication protocols and information types require application support.',
    ],
  },
  {
    id: 'fields',
    title: 'Map the response fields',
    paragraphs: [
      'A field location says where to find a value in the provider response. /collections means the collections list. /name means the name inside each collection. /result selects a nested result section.',
      'Map the list once. NodeHerder applies the same name, dates and colour fields to every returned collection, including new colours. You do not need a configuration section per bin.',
      'These locations use JSON Pointer: a slash separates each level. If the provider moves a field into another object, update its location, for example /details/name. This is a small mapping of the fields we use, not a copy of the provider’s entire schema.',
      'Dates are calendar dates. A missing date or an empty list does not prove a collection was cancelled. Approximate or neighbouring-address results are rejected in this release.',
    ],
  },
  {
    id: 'repair',
    title: 'Repair a changed URL or response',
    paragraphs: [
      'For a changed provider address, edit Provider URL. For a changed endpoint, open its request and edit the URL path or parameters. For a renamed response field, update its field location under Collection fields or Values saved from the response.',
      'Validate checks the configuration without sending a request. Save changes applies it. Then open an existing integration and choose Refresh to check the provider. A failed refresh keeps the last saved dates.',
      'Changing a destination requires review before saved inputs or an existing token can be sent there. The YAML itself cannot approve a credential transfer. If a new login method or a different response transformation is required, the application may need an update.',
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
export function helpTopics(search: string) {
  const q = search.trim().toLowerCase();
  return CONNECTION_HELP.filter((t) => !q || `${t.title} ${t.paragraphs.join(' ')}`.toLowerCase().includes(q));
}
