# Projet GO – Traitement d'images (séquentiel vs parallèle)

Petit projet pour comparer des versions séquentielles et parallèles d'algorithmes sur des images représentées en matrices de `Pixel`.

## Prérequis
- Go installé (version récente recommandée)
- Une image d'entrée présente à la racine du projet sous le nom `image.jpg` (un exemple est déjà fourni)

## Lancer le programme
Exécute le pipeline actuel (extraction des pixels, comparaison séquentiel/parallèle, traitement et export PNG).

```powershell
cd "C:\Users\Hector\Desktop\INSA Lyon\3A-TC\S1\ELP-GO Projet\projet-go"
# Optionnel mais conseillé pour récupérer les dépendances si besoin
go mod tidy
# Lancer
go run .
```

Sortie attendue (exemple) : dimensions, temps séquentiel/parallèle, speedup, et génération de `out.png`.

## Benchmarks
Benchmarks Go comparant séquentiel vs parallèle pour l'extraction et le passage en niveaux de gris.

```powershell
go test -bench=. -benchmem
```

Astuces pour stabiliser les chiffres :
- Multiplier les itérations et runs :
  ```powershell
  go test -bench=. -benchmem -benchtime=2s -count=3
  ```
- Fixer le nombre de CPU utilisés (facultatif) :
  ```powershell
  $env:GOMAXPROCS = 8
  go test -bench=. -benchmem
  ```

## Structure
- [main.go](main.go) : point d'entrée, affiche la comparaison et écrit `out.png`.
- [nonparallel.go](nonparallel.go) : fonctions séquentielles et utilitaires (chargement/sauvegarde, `pixelsToImage`, `downscalePixels`).
- [parallel.go](parallel.go) : fonctions parallélisées (`extractPixelsParallel`, `blackWhiteParallel`).
- [bench_test.go](bench_test.go) : benchmarks séquentiel vs parallèle.
- [halo.go](halo.go) : fichier neutralisé après refactorisation (la logique a été répartie).

## Notes
- Les fonctions parallèles découpent l'image par bandes de lignes et utilisent `runtime.NumCPU()` workers.
- Les versions séquentielles restent la référence pour valider la justesse des résultats.
- Les performances dépendent de la taille de l'image, du nombre de cœurs et de la charge système.

## Résultats
Le programme affiche un speedup estimé entre les deux versions. Pour des images petites, le surcoût de parallélisation peut dominer; pour des images grandes (HD/4K), le gain est généralement net.

## Remap de pixels (sans changer les couleurs)
- Place les pixels de l'image source pour matcher la distribution de couleurs de l'image cible.
- Pas de modification de valeurs : seules les positions changent.
- Pré-requis : une image `target.jpg` avec les mêmes dimensions que `image.jpg`.
- Le remap est déclenché automatiquement si `target.jpg` existe, et produit `remap.png`.
