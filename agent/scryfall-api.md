# Scryfall API Reference

## Base URL

```
https://api.scryfall.com
```

- HTTPS only (TLS 1.2+), no HTTP
- UTF-8 encoding throughout
- Rate limit for search: **2 requests/second (500ms between requests)**

## Required Headers

All requests **must** include:

| Header | Value |
|---|---|
| `User-Agent` | Name of your app, e.g. `MTGAlternatives/1.0` |
| `Accept` | Any acceptable, e.g. `application/json` or `*/*` |

Without these headers, Scryfall returns HTTP 400 with an empty body.

## Error Responses

Scryfall errors are JSON:

```json
{
  "object": "error",
  "code": "not_found",
  "status": 404,
  "details": "No cards found matching your query."
}
```

Key fields:
- `details` (string) — human-readable error message
- `status` (int) — HTTP status code
- `code` (string) — machine-readable error code (e.g., `not_found`, `bad_request`)

## Card Search

```
GET /cards/search?q={query}
```

### Parameters

| Parameter | Type | Required | Description |
|---|---|---|---|
| `q` | String | Yes | Fulltext search query (max 1000 chars). Must be URL-encoded. |
| `unique` | String | No | Rollup mode: `cards` (default), `art`, `prints` |
| `order` | String | No | Sort: `name` (default), `set`, `released`, `rarity`, `color`, `usd`, `tix`, `eur`, `cmc`, `power`, `toughness`, `edhrec`, `penny`, `artist`, `review` |
| `dir` | String | No | Sort direction: `auto` (default), `asc`, `desc` |
| `page` | Integer | No | Page number, default 1 |
| `format` | String | No | `json` (default) or `csv` |
| `include_extras` | Boolean | No | Include tokens, planes, etc. Default false. |
| `include_multilingual` | Boolean | No | Include all languages. Default false. |
| `include_variations` | Boolean | No | Include rare variants. Default false. |

### Response

Returns a **List object** (not an array):

```json
{
  "object": "list",
  "total_cards": 2,
  "has_more": false,
  "data": [ ... cards ... ]
}
```

| Field | Type | Description |
|---|---|---|
| `object` | string | Always `"list"` |
| `total_cards` | int | Total matching cards across all pages |
| `has_more` | boolean | `true` if more pages exist |
| `data` | array | Array of Card objects (up to 175 per page) |
| `next_page` | string | (Present if `has_more` is true) URL of next page |

**Important:** The cards array is in `data`, not `cards`.

### Card Object (fields used by this project)

```json
{
  "object": "card",
  "id": "bd8fa327-dd41-4737-8f19-2cf5eb1f7cdd",
  "name": "Black Lotus",
  "image_uris": {
    "small": "https://cards.scryfall.io/small/...",
    "normal": "https://cards.scryfall.io/normal/...",
    "large": "https://cards.scryfall.io/large/...",
    "png": "https://cards.scryfall.io/png/...",
    "art_crop": "https://cards.scryfall.io/art_crop/...",
    "border_crop": "https://cards.scryfall.io/border_crop/..."
  }
}
```

The `id` field is the Scryfall UUID used to uniquely identify a card print.

### Double-Faced Cards

Some cards (transform, modal_dfc) have **no top-level `image_uris`**. Instead, images are in `card_faces[].image_uris`. The current code skips these cards.

### Search Query Syntax

Scryfall uses the same fulltext search as the website:

| Syntax | Example | Meaning |
|---|---|---|
| Plain text | `black lotus` | Name contains words |
| `c:{color}` | `c:red` | Color identity |
| `t:{type}` | `t:creature` | Card type |
| `cmc:{n}` | `cmc=3` | Mana value |
| `pow:{n}` | `pow=3` | Power |
| `tou:{n}` | `tou=3` | Toughness |
| `e:{set}` | `e:ltr` | Set code |
| `f:{format}` | `f:commander` | Legal in format |
| `r:{rarity}` | `r:mythic` | Rarity |
| `o:{text}` | `o:"draw a card"` | Oracle text contains |

### Pagination

First page: `GET /cards/search?q=...`
Subsequent pages: follow `next_page` URL. Each page has up to 175 cards.

## Image URLs

Card images are hosted at `https://cards.scryfall.io/`. Available sizes:

| Size | Description |
|---|---|
| `small` | 146 × 204 px |
| `normal` | 488 × 680 px (standard display) |
| `large` | 672 × 936 px |
| `png` | 745 × 1040 px, full quality |
| `art_crop` | Cropped artwork only |
| `border_crop` | Cropped to border edge |

## Image Usage Rules

- Do not cover, crop, or clip copyright/artist name
- Do not distort, stretch, or color-shift images
- Do not add watermarks
- When using `art_crop`, list artist name and copyright nearby

## Fulltext Search Operator Reference

Full syntax: https://scryfall.com/docs/syntax
