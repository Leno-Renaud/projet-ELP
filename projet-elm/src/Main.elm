module Main exposing (..)

import Browser
import Browser.Events exposing (onKeyDown)
import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (onClick)
import Http
import Json.Decode as Decode
import Random


-- MAIN


main : Program () Model Msg
main = Browser.element
        {init = init,
        view = view,
        update = update,
        subscriptions = subscriptions
        }



-- MODEL


type GameState    -- Déclaration du type GameState
    = Playing
    | Won
    | Lost


type alias Model =
    {guesses : List String,
    currentGuess : String,
    targetWord : String,
    gameState : GameState,
    validWords : List String,
    errorMessage : Maybe String
    }


initialValidWords : List String
initialValidWords = [ "WORLD" ]


init : () -> ( Model, Cmd Msg )
init _ =
    ( {guesses = [],
      currentGuess = "",
      targetWord = "WORLD",
      gameState = Playing,
      validWords = initialValidWords,
      errorMessage = Nothing
      },
    Http.get
        { url = "/words.txt",
        expect = Http.expectString GotWords
        }
    )


randomWordGenerator : List String -> Random.Generator String
randomWordGenerator words =
    case words of
        [] ->
            Random.constant "WORLD"

        first :: rest ->
            Random.uniform first rest



-- UPDATE


type Msg 
    = KeyPressed String
    | NewTargetWord String
    | Restart
    | GotWords (Result Http.Error String)


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        GotWords result ->  -- Gestion de la réponse HTTP pour charger les mots valides
            case result of
                Ok fullText -> -- Si la requête réussit, on traite le texte reçu
                    let
                        words =
                            fullText --> on diviser le texte en mots, les mettre en majuscules et filtrer ceux de longueur 5
                                |> String.words
                                |> List.map String.toUpper
                                |> List.filter (\w -> String.length w == 5)
                    in
                    ( { model | validWords = words } -- Mise à jour du modèle avec les mots valides
                    , Random.generate NewTargetWord (randomWordGenerator words) -- Génération d'un mot cible aléatoire
                    )

                Err error ->
                    ( { model | errorMessage = Just "Failed to load words" }, Cmd.none )

        Restart ->
            ( { model
                | guesses = []
                , currentGuess = ""
                , gameState = Playing
              }
            , Random.generate NewTargetWord (randomWordGenerator model.validWords)
            )

        NewTargetWord word ->
            ( { model | targetWord = word }, Cmd.none )

        KeyPressed key ->
            if model.gameState /= Playing then
                ( model, Cmd.none )

            else
                handleKey key model


handleKey : String -> Model -> ( Model, Cmd Msg )-- Gestion des entrées clavier
handleKey key model =
    if key == "Enter" then
        submitGuess model

    else if key == "Backspace" then
        ( { model | currentGuess = String.dropRight 1 model.currentGuess }, Cmd.none )  

    else if String.length key == 1 && String.all Char.isAlpha key then-- Si la touche est une lettre
        if String.length model.currentGuess < 5 then
            ( { model | currentGuess = model.currentGuess ++ String.toUpper key }, Cmd.none )

        else
            ( model, Cmd.none )

    else
        ( model, Cmd.none )


submitGuess : Model -> ( Model, Cmd Msg )
submitGuess model =
    let
        guess =
            model.currentGuess
    in
    if String.length guess /= 5 then
        ( model, Cmd.none )

    else if not (List.member guess model.validWords) then
         ( model, Cmd.none )
    else
        let
            newGuesses =-- Mise à jour de la liste des suppositions avec la nouvelle supposition
                model.guesses ++ [ guess ]

            newGameState =
                if guess == model.targetWord then
                    Won

                else if List.length newGuesses >= 6 then
                    Lost

                else
                    Playing
        in
        ( { model
            | guesses = newGuesses
            , currentGuess = ""
            , gameState = newGameState
          }
        , Cmd.none
        )



-- SUBSCRIPTIONS


subscriptions : Model -> Sub Msg
subscriptions _ =
    onKeyDown (Decode.map KeyPressed (Decode.field "key" Decode.string))



-- VIEW


view : Model -> Html Msg
view model =
    div [ style "display" "flex", style "flex-direction" "column", style "align-items" "center", style "font-family" "sans-serif", style "padding-top" "20px" ]
        [ h1 [] [ text "Elm Wordle" ]
        , div [ style "font-size" "12px", style "color" "#666", style "margin-bottom" "10px" ]
            [ text (case model.errorMessage of
                Just err -> "Error: " ++ err
                Nothing -> "Words loaded: " ++ String.fromInt (List.length model.validWords))
            ]
        , viewGrid model
        , viewMessage model
        , if model.gameState /= Playing then
            button 
                [ onClick Restart
                , style "padding" "10px 20px"
                , style "font-size" "16px"
                , style "margin-top" "20px" 
                , style "cursor" "pointer"
                ] 
                [ text "Restart" ]
          else
            div [ style "height" "42px" ] [] -- Placeholder
        ]


viewMessage : Model -> Html msg
viewMessage model =
    case model.gameState of
        Playing ->
             div [ style "height" "20px" ] []

        Won ->
            div [ style "margin" "10px", style "color" "green", style "font-weight" "bold" ] [ text "You Won!" ]

        Lost ->
            div [ style "margin" "10px", style "color" "red", style "font-weight" "bold" ] [ text ("Game Over! The word was " ++ model.targetWord) ]


viewGrid : Model -> Html msg
viewGrid model =
    let
        rows = List.range 0 5-- Six lignes pour les suppositions
    in
    div [ style "display" "grid", style "grid-template-rows" "repeat(6, 1fr)", style "gap" "5px" ]
        (List.map (viewRow model) rows)


viewRow : Model -> Int -> Html msg
viewRow model rowIndex =
    let
        guessesCount = List.length model.guesses
        
        (word, isSubmitted) =
            if rowIndex < guessesCount then
                ( Maybe.withDefault "" (List.head (List.drop rowIndex model.guesses)), True )
            else if rowIndex == guessesCount then
                ( model.currentGuess, False )
            else
                ( "", False )
                
        chars = padList 5 ' ' (String.toList word)
    in
    div [ style "display" "grid", style "grid-template-columns" "repeat(5, 1fr)", style "gap" "5px" ]
        (List.indexedMap (viewCell model.targetWord isSubmitted word) chars)


padList : Int -> a -> List a -> List a
padList n default list =
    let
        len = List.length list
    in
    if len < n then
        list ++ List.repeat (n - len) default
    else
        list


viewCell : String -> Bool -> String -> Int -> Char -> Html msg
viewCell targetWord isSubmitted guessWord index char =
    let
        content =
            if char == ' ' then "" else String.fromChar char

        bgColor =
            if isSubmitted then
                getCellColor targetWord guessWord index
            else
                "white"
        
        textColor =
            if isSubmitted then "white" else "black"
            
        borderColor =
             if not isSubmitted && char /= ' ' then "#888" else "#ccc"

    in
    div
        [ style "width" "60px"
        , style "height" "60px"
        , style "border" ("2px solid " ++ borderColor)
        , style "display" "flex"
        , style "justify-content" "center"
        , style "align-items" "center"
        , style "font-size" "32px"
        , style "font-weight" "bold"
        , style "text-transform" "uppercase"
        , style "background-color" bgColor
        , style "color" textColor
        ]
        [ text content ]


getCellColor : String -> String -> Int -> String
getCellColor target guess index =
    let
        targetList = String.toList target-- Liste des caractères du mot cible
        guessList = String.toList guess
        
        guessChar = -- Caractère de la supposition à l'index donné
            guessList 
            |> List.drop index 
            |> List.head 
            |> Maybe.withDefault ' '
            
        targetChar = -- Caractère du mot cible à l'index donné
            targetList 
            |> List.drop index 
            |> List.head 
            |> Maybe.withDefault ' '
    in
    if guessChar == targetChar then
        "#6aaa64" -- Green
    else
        if List.member guessChar targetList then
             "#c9b458" -- Yellow
        else
            "#787c7e" -- Gray

