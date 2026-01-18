// Package selector provides pattern selection for MessageFormat 2.0
// TypeScript original code: select-pattern.ts module
package selector

import (
	"github.com/kaptinlin/messageformat-go/internal/resolve"
	"github.com/kaptinlin/messageformat-go/pkg/datamodel"
	"github.com/kaptinlin/messageformat-go/pkg/errors"
	"github.com/kaptinlin/messageformat-go/pkg/logger"
)

// SelectPattern selects the appropriate pattern from a message
// TypeScript original code:
//
//	export function selectPattern(context: Context, message: Message): Pattern {
//	  switch (message.type) {
//	    case 'message':
//	      return message.pattern;
//
//	    case 'select': {
//	      const ctx = message.selectors.map(sel => {
//	        const selector = resolveVariableRef(context, sel);
//	        let selectKey;
//	        if (typeof selector.selectKey === 'function') {
//	          selectKey = selector.selectKey.bind(selector);
//	        } else {
//	          context.onError(new MessageSelectionError('bad-selector'));
//	          selectKey = () => null;
//	        }
//	        return {
//	          selectKey,
//	          best: null as string | null,
//	          keys: null as Set<string> | null
//	        };
//	      });
//
//	      let candidates = message.variants;
//	      loop: for (let i = 0; i < ctx.length; ++i) {
//	        const sc = ctx[i];
//	        if (!sc.keys) {
//	          sc.keys = new Set();
//	          for (const { keys } of candidates) {
//	            const key = keys[i];
//	            if (!key) break loop; // key-mismatch error
//	            if (key.type !== '*') sc.keys.add(key.value);
//	          }
//	        }
//	        try {
//	          sc.best = sc.keys.size ? sc.selectKey(sc.keys) : null;
//	        } catch (error) {
//	          context.onError(new MessageSelectionError('bad-selector', error));
//	          sc.selectKey = () => null;
//	          sc.best = null;
//	        }
//
//	        // Leave out all candidate variants that aren't the best,
//	        // or only the catchall ones, if nothing else matches.
//	        candidates = candidates.filter(v => {
//	          const k = v.keys[i];
//	          if (k.type === '*') return sc.best == null;
//	          return sc.best === k.value;
//	        });
//
//	        // If we've run out of candidates,
//	        // drop the previous best key of the preceding selector,
//	        // reset all subsequent key sets,
//	        // and restart the loop.
//	        if (candidates.length === 0) {
//	          if (i === 0) break; // No match; should not happen
//	          const prev = ctx[i - 1];
//	          if (prev.best == null) prev.keys?.clear();
//	          else prev.keys?.delete(prev.best);
//	          for (let j = i; j < ctx.length; ++j) ctx[j].keys = null;
//	          candidates = message.variants;
//	          i = -1;
//	        }
//	      }
//
//	      const res = candidates[0];
//	      if (!res) {
//	        // This should not be possible with a valid message.
//	        context.onError(new MessageSelectionError('no-match'));
//	        return [];
//	      }
//	      return res.value;
//	    }
//
//	    default:
//	      context.onError(new MessageSelectionError('bad-selector'));
//	      return [];
//	  }
//	}
func SelectPattern(context *resolve.Context, message datamodel.Message) datamodel.Pattern {
	// matches TypeScript: switch (message.type)
	switch msg := message.(type) {
	case *datamodel.PatternMessage:
		// matches TypeScript: case 'message': return message.pattern;
		return msg.Pattern()

	case *datamodel.SelectMessage:
		// matches TypeScript: case 'select': { ... }
		return selectVariantPattern(context, msg)

	default:
		// matches TypeScript: default: context.onError(new MessageSelectionError('bad-selector')); return [];
		logger.Error("unsupported message type for pattern selection", "type", message.Type())
		if context.OnError != nil {
			context.OnError(errors.NewMessageSelectionError(
				errors.ErrorTypeBadSelector,
				nil,
			))
		}
		return datamodel.NewPattern(nil)
	}
}

// selectorContext represents the context for a single selector
// TypeScript original code:
//
//	{
//	  selectKey,
//	  best: null as string | null,
//	  keys: null as Set<string> | null
//	}
type selectorContext struct {
	selectKey func(map[string]bool) *string // matches TypeScript selectKey function
	best      *string                       // matches TypeScript: best: null as string | null
	keys      map[string]bool               // matches TypeScript: keys: null as Set<string> | null
}

// selectVariantPattern selects the best matching variant pattern
// TypeScript original code: select case logic in selectPattern function
func selectVariantPattern(context *resolve.Context, msg *datamodel.SelectMessage) datamodel.Pattern {
	selectors := msg.Selectors()
	variants := msg.Variants()

	// matches TypeScript: const ctx = message.selectors.map(sel => { ... });
	selectorCtxs := make([]*selectorContext, len(selectors))
	for i, selector := range selectors {
		// matches TypeScript: const selector = resolveVariableRef(context, sel);
		mv := resolve.ResolveVariableRef(context, &selector)

		var selectKeyFunc func(map[string]bool) *string
		// matches TypeScript: if (typeof selector.selectKey === 'function')
		// Check if the MessageValue supports selection by testing with a dummy key
		testKeys := []string{"test"}
		if _, err := mv.SelectKeys(testKeys); err == nil {
			// matches TypeScript: selectKey = selector.selectKey.bind(selector);
			selectKeyFunc = func(availableKeys map[string]bool) *string {
				// Convert map to slice for selection
				keySlice := make([]string, 0, len(availableKeys))
				for key := range availableKeys {
					keySlice = append(keySlice, key)
				}

				if len(keySlice) == 0 {
					return nil
				}

				// Call the MessageValue's SelectKeys method
				selectedKeys, err := mv.SelectKeys(keySlice)
				if err != nil || len(selectedKeys) == 0 {
					return nil
				}

				// Return the first selected key
				return &selectedKeys[0]
			}
		} else {
			// matches TypeScript: context.onError(new MessageSelectionError('bad-selector')); selectKey = () => null;
			if context.OnError != nil {
				context.OnError(errors.NewMessageSelectionError(
					errors.ErrorTypeBadSelector,
					nil,
				))
			}
			selectKeyFunc = func(map[string]bool) *string { return nil }
		}

		// matches TypeScript: return { selectKey, best: null as string | null, keys: null as Set<string> | null };
		selectorCtxs[i] = &selectorContext{
			selectKey: selectKeyFunc,
		}
	}

	// matches TypeScript: let candidates = message.variants;
	candidates := variants

	// matches TypeScript: loop: for (let i = 0; i < ctx.length; ++i) {
	for i := 0; i < len(selectorCtxs); i++ {
		sc := selectorCtxs[i]

		// matches TypeScript: if (!sc.keys) { sc.keys = new Set(); ... }
		if sc.keys == nil {
			sc.keys = make(map[string]bool)
			// matches TypeScript: for (const { keys } of candidates) { const key = keys[i]; ... }
			for _, variant := range candidates {
				keys := variant.Keys()
				// matches TypeScript: if (!key) break loop; // key-mismatch error
				if i >= len(keys) {
					goto loopEnd // equivalent to break loop in TypeScript
				}
				key := keys[i]
				// matches TypeScript: if (key.type !== '*') sc.keys.add(key.value);
				if !datamodel.IsCatchallKey(key) {
					if literal, ok := key.(*datamodel.Literal); ok {
						sc.keys[literal.Value()] = true
					}
				}
			}
		}

		// matches TypeScript: try { sc.best = sc.keys.size ? sc.selectKey(sc.keys) : null; } catch (error) { ... }
		func() {
			defer func() {
				if r := recover(); r != nil {
					// matches TypeScript: context.onError(new MessageSelectionError('bad-selector', error));
					if context.OnError != nil {
						context.OnError(errors.NewMessageSelectionError(
							errors.ErrorTypeBadSelector,
							nil,
						))
					}
					// matches TypeScript: sc.selectKey = () => null; sc.best = null;
					sc.selectKey = func(map[string]bool) *string { return nil }
					sc.best = nil
				}
			}()

			// matches TypeScript: sc.best = sc.keys.size ? sc.selectKey(sc.keys) : null;
			if len(sc.keys) > 0 {
				sc.best = sc.selectKey(sc.keys)
			}
		}()

		// matches TypeScript: candidates = candidates.filter(v => { ... });
		var newCandidates []datamodel.Variant
		for _, variant := range candidates {
			keys := variant.Keys()
			if i >= len(keys) {
				continue
			}

			key := keys[i]
			// matches TypeScript: if (k.type === '*') return sc.best == null;
			if datamodel.IsCatchallKey(key) {
				if sc.best == nil {
					newCandidates = append(newCandidates, variant)
				}
			} else {
				// matches TypeScript: return sc.best === k.value;
				if literal, ok := key.(*datamodel.Literal); ok {
					if sc.best != nil && *sc.best == literal.Value() {
						newCandidates = append(newCandidates, variant)
					}
				}
			}
		}

		candidates = newCandidates

		// matches TypeScript: if (candidates.length === 0) { ... }
		if len(candidates) == 0 {
			// matches TypeScript: if (i === 0) break; // No match; should not happen
			if i == 0 {
				break
			}

			// matches TypeScript: const prev = ctx[i - 1]; if (prev.best == null) prev.keys?.clear(); else prev.keys?.delete(prev.best);
			prev := selectorCtxs[i-1]
			if prev.best == nil {
				prev.keys = nil // equivalent to clear()
			} else {
				delete(prev.keys, *prev.best)
			}

			// matches TypeScript: for (let j = i; j < ctx.length; ++j) ctx[j].keys = null;
			for j := i; j < len(selectorCtxs); j++ {
				selectorCtxs[j].keys = nil
			}

			// matches TypeScript: candidates = message.variants; i = -1;
			candidates = variants
			i = -1 // Will be incremented to 0 in next iteration
		}
	}

loopEnd:
	// matches TypeScript: const res = candidates[0];
	if len(candidates) > 0 {
		res := candidates[0]
		// matches TypeScript: return res.value;
		return res.Value()
	}

	// matches TypeScript: if (!res) { context.onError(new MessageSelectionError('no-match')); return []; }
	if context.OnError != nil {
		context.OnError(errors.NewMessageSelectionError(
			errors.ErrorTypeNoMatch,
			nil,
		))
	}

	return datamodel.NewPattern(nil)
}
