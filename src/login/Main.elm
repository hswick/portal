import Browser
import Html
import Html.Styled exposing (..)
import Html.Styled.Attributes exposing (..)
import Html.Styled.Events exposing (..)
import Http
import Json.Decode as Decode
import Json.Encode as Encode
import Browser.Navigation as Nav


main =
    Browser.element
        { init = init
        , update = update
        , subscriptions = always Sub.none
        , view = view >> toUnstyled
        }

        
type alias ActiveUser =
    { id : Int
    , accessToken : String
    }

    
type alias Model =
    { loginUsernameText : String
    , loginPasswordText : String
    , errorMessage : String
    }


init : () -> ( Model, Cmd Msg )
init _ =
    ( { loginUsernameText = ""
      , loginPasswordText = ""
      , errorMessage = ""
      }
    , Cmd.none )


postLogin : String -> String -> Cmd Msg
postLogin username password =
          Http.post
                { url = "/login/credentials"
                , body = Http.jsonBody (credentialsEncoder username password)
                , expect = Http.expectJson PostLogin activeUserDecoder
                }

              
activeUserDecoder : Decode.Decoder ActiveUser
activeUserDecoder =
    Decode.map2 ActiveUser
        (Decode.field "id" Decode.int)
        (Decode.field "accessToken" Decode.string)
              

credentialsEncoder : String -> String -> Encode.Value
credentialsEncoder username password =
             Encode.object
                 [ ("username", Encode.string username)
                 , ("password", Encode.string password)
                 ]


-- UPDATE


type Msg
     = LoginUsernameInput String
     | LoginPasswordInput String
     | SubmitLogin
     | PostLogin (Result Http.Error ActiveUser)


activeUserToUrl : ActiveUser -> String
activeUserToUrl au =
        ("/welcome?user_id=" ++ (String.fromInt au.id) ++ "&access_token=" ++ au.accessToken)
            
        
update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
       case msg of
            LoginUsernameInput username ->
                              ( { model | loginUsernameText = username }, Cmd.none )

            LoginPasswordInput password ->
                              ( { model | loginPasswordText = password }, Cmd.none )

            SubmitLogin ->
                        ( model, postLogin model.loginUsernameText model.loginPasswordText )

            PostLogin result ->
                      case result of
                           Ok activeUser ->
                              ( model, Nav.load (activeUserToUrl activeUser) )

                           Err _ ->
                               ( { model | errorMessage = "An error has occurred" }, Cmd.none )


-- VIEW


view : Model -> Html Msg
view model =
     div []
         [ loginView model
         , text model.errorMessage
         ]


loginView : Model -> Html Msg
loginView model =
     div []
         [ input [ onInput LoginUsernameInput, placeholder "Username", value model.loginUsernameText ] []
         , input [ onInput LoginPasswordInput, placeholder "Password", value model.loginPasswordText ] []
         , button [ onClick SubmitLogin ] [ text "Login" ]
         ]
