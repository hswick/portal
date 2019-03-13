import Browser
import Browser.Navigation as Nav
import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (..)
import Url
import Http
import Json.Decode as Decode
import Json.Encode as Encode


-- MAIN


main : Program Flags Model Msg
main =
  Browser.application
    { init = init
    , view = view
    , update = update
    , subscriptions = subscriptions
    , onUrlChange = UrlChanged
    , onUrlRequest = LinkClicked
    }


-- MODEL


type alias Model =
  { key : Nav.Key
  , url : Url.Url
  , otherUsernameText : String
  , otherPasswordText : String
  , changePasswordText : String
  , changeUsernameText : String
  , newPassword : String
  , adminChecked : Bool
  , id : Int
  , accessToken : String
  }


type alias Flags =
    { id : Int
    , accessToken : String
    }
    

init : Flags -> Url.Url -> Nav.Key -> ( Model, Cmd Msg )
init flags url key =
  ( { key = key
    , url = url
    , otherUsernameText = ""
    , otherPasswordText = ""
    , changePasswordText = ""
    , changeUsernameText = ""
    , newPassword = ""
    , adminChecked = False
    , id = flags.id
    , accessToken = flags.accessToken
    }
  , Cmd.none )

    
postChangeUsername : Model -> Cmd Msg
postChangeUsername model =
    Http.post
        { url = "/update/username"
        , body = Http.jsonBody (updateUsernameEncoder model)
        , expect = Http.expectWhatever PostChangeUsername
        }

        
updateUsernameEncoder : Model -> Encode.Value
updateUsernameEncoder model =
    Encode.object
        [ ("username", Encode.string model.changeUsernameText)
        , ("id", Encode.int model.id)
        , ("access_token", Encode.string model.accessToken)
        ]

postChangePassword : Model -> Cmd Msg
postChangePassword model =
    Http.post
        { url = "/update/password"
        , body = Http.jsonBody (updatePasswordEncoder model)
        , expect = Http.expectWhatever PostChangeUsername
        }

        
updatePasswordEncoder : Model -> Encode.Value
updatePasswordEncoder model =
    Encode.object
        [ ("password", Encode.string model.changePasswordText)
        , ("id", Encode.int model.id)
        , ("access_token", Encode.string model.accessToken)
        ]
        

postNewUser : Model -> Cmd Msg
postNewUser model =
    Http.post
        { url = "/register/credentials"
        , body = Http.jsonBody (newUserEncoder model)
        , expect = Http.expectWhatever PostAdminAction
        }

        
newUserEncoder : Model -> Encode.Value
newUserEncoder model =
    let
        admin =
            if model.adminChecked then
                "true"
            else
                "false"
    in
        Encode.object
            [ ("username", Encode.string model.otherUsernameText)
            , ("password", Encode.string model.otherPasswordText)
            , ("id", Encode.int model.id)
            , ("admin", Encode.string admin)
            ]


postNewPassword : Model -> Cmd Msg
postNewPassword model =
    Http.post
        { url = "/admin/password"
        , body = Http.jsonBody (adminActionEncoder model)
        , expect = Http.expectJson PostNewPassword newPasswordDecoder
        }
        
        
type alias NewPasswordBody =
    { password : String }

        
newPasswordDecoder : Decode.Decoder NewPasswordBody
newPasswordDecoder =
    Decode.map NewPasswordBody
        (Decode.field "password" Decode.string)


postMakeAdmin : Model -> Cmd Msg
postMakeAdmin model =
    Http.post
        { url = "/admin/new"
        , body = Http.jsonBody (adminActionEncoder model)
        , expect = Http.expectWhatever PostAdminAction
        }


postRevokeAdmin : Model -> Cmd Msg
postRevokeAdmin model =
    Http.post
        { url = "/admin/revoke"
        , body = Http.jsonBody (adminActionEncoder model)
        , expect = Http.expectWhatever PostAdminAction
        }


postDeleteUser : Model -> Cmd Msg
postDeleteUser model =
    Http.post
        { url = "/admin/delete/user"
        , body = Http.jsonBody (adminActionEncoder model)
        , expect = Http.expectWhatever PostAdminAction
        }


adminActionEncoder : Model -> Encode.Value
adminActionEncoder model =
    Encode.object
        [ ("username", Encode.string model.otherUsernameText)
        , ("access_token", Encode.string model.accessToken)
        , ("id", Encode.int model.id)
        ]
        
        
-- UPDATE


type Msg
  = LinkClicked Browser.UrlRequest
  | UrlChanged Url.Url
  | OtherUsernameInput String
  | OtherPasswordInput String
  | ToggleAdmin
  | NewUser
  | AdminNewPassword
  | MakeAdmin
  | RevokeAdmin
  | DeleteUser
  | ChangeUsernameInput String
  | ChangePasswordInput String
  | SubmitChangeUsername
  | SubmitChangePassword
  | PostChangeUsername (Result Http.Error ())
  | PostChangePassword (Result Http.Error ())
  | PostAdminAction (Result Http.Error ())
  | PostNewPassword (Result Http.Error NewPasswordBody)



update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
  case msg of
    LinkClicked urlRequest ->
      case urlRequest of
        Browser.Internal url ->
          ( model, Nav.pushUrl model.key (Url.toString url) )

        Browser.External href ->
          ( model, Nav.load href )

    UrlChanged url ->
      ( { model | url = url }, Cmd.none )

    ChangeUsernameInput username ->
        ( { model | changeUsernameText = username }, Cmd.none )

    ChangePasswordInput password ->
        ( { model | changePasswordText = password }, Cmd.none )

    SubmitChangeUsername ->
        ( model, postChangeUsername model )

    SubmitChangePassword ->
        ( model, postChangePassword model )

    OtherUsernameInput username ->
        ( { model | otherUsernameText = username }, Cmd.none )

    OtherPasswordInput password ->
        ( { model | otherPasswordText = password }, Cmd.none )

    NewUser ->
        ( model, postNewUser model )

    ToggleAdmin ->
        ( { model | adminChecked = not model.adminChecked }, Cmd.none )

    AdminNewPassword ->
        ( model, postNewPassword model )

    MakeAdmin ->
        ( model, postMakeAdmin model )

    RevokeAdmin ->
        ( model, postRevokeAdmin model )

    DeleteUser ->
        ( model, postDeleteUser model )

    PostChangeUsername _ ->
        ( model, Cmd.none )

    PostChangePassword _ ->
        ( model, Cmd.none )

    PostNewPassword result ->
        case result of
            Ok object ->
                ( { model | newPassword = object.password }, Cmd.none )
            Err _ ->
                ( model, Cmd.none )

    PostAdminAction _ ->
        ( model, Cmd.none )


-- SUBSCRIPTIONS


subscriptions : Model -> Sub Msg
subscriptions _ =
  Sub.none


-- VIEW
    

view : Model -> Browser.Document Msg
view model =
  { title = "URL Interceptor"
  , body = [ viewRouter model ]
  }


viewRouter : Model -> Html Msg
viewRouter model =
    case (Url.toString model.url) of
        "/settings" ->
            settingsView model

        _ ->
            welcomeView model
                

viewLink : String -> Html Msg
viewLink path =
  li [] [ a [ href path ] [ text path ] ]


welcomeView : Model -> Html Msg
welcomeView model =
    table []
        [ tr []
              [ td [] [ viewLink "/settings" ]
              , td [] [ a [ href "/canban" ] [ text "canban" ] ]
              ]
         ]


settingsView : Model -> Html Msg
settingsView model =
    div []
        [ changeUsernameView model
        , changePasswordView model
        , adminSettingsView model
        ]

        
changeUsernameView : Model -> Html Msg
changeUsernameView model =
    div []
        [ input [ onInput ChangeUsernameInput, placeholder "New Username", value model.changeUsernameText ] []
        , button [ onClick SubmitChangeUsername ] [ text "Submit" ]
        ]

        
changePasswordView : Model -> Html Msg
changePasswordView model =
    div []
        [ input [ onInput ChangePasswordInput, placeholder "New Password", value model.changePasswordText ] []
        , button [ onClick SubmitChangePassword ] [ text "Submit" ]
        ]


adminSettingsView : Model -> Html Msg
adminSettingsView model =
    div []
        [ input [ onInput OtherUsernameInput, placeholder "Other username", value model.otherUsernameText ] []
        , button [ onClick AdminNewPassword ] [ text "New Password" ]
        , text model.newPassword
        , button [ onClick MakeAdmin ] [ text "Make Admin" ]
        , button [ onClick RevokeAdmin ] [ text "Revoke Admin" ]
        , button [ onClick DeleteUser ] [ text "Delete User" ]
        , newUserView model
        ]


newUserView : Model -> Html Msg
newUserView model =
        div []
            [ input [ onInput OtherUsernameInput, placeholder "Other username", value model.otherUsernameText ] []
            , input [ onInput OtherPasswordInput, placeholder "New user's password", value model.otherPasswordText ] []
            , input [ type_ "checkbox", checked model.adminChecked, onClick ToggleAdmin ] []
            , button [ onClick NewUser ] [ text "New User" ]
            ]
