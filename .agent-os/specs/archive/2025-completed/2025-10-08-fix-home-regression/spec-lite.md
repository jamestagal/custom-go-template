# Spec Summary (Lite)

Fix home.html regression on global-store-system branch where UserProfile component functions are stripped during fence parsing with store support. The `ParseFenceContentWithStores()` function removes function definitions (formatDate, getRoleBadge) causing undefined errors and empty component cards. Investigation and fix must preserve all fence functions when processing store imports.
