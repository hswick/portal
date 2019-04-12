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
  , newUserUsernameText : String
  , newUserPasswordText : String
  , changePasswordText : String
  , changeUsernameText : String
  , oldPasswordText : String
  , newPassword : String
  , adminChecked : Bool
  , id : Int
  , accessToken : String
  , name : String
  , apps : List String
  , admin : Bool
  }


type alias Flags =
    { name : String
    , id : Int
    , admin : Bool
    , accessToken : String
    , apps : List String
    }


init : Flags -> Url.Url -> Nav.Key -> ( Model, Cmd Msg )
init flags url key =
  ( { key = key
    , url = url
    , otherUsernameText = ""
    , newUserUsernameText = ""
    , newUserPasswordText = ""
    , changePasswordText = ""
    , changeUsernameText = ""
    , oldPasswordText = ""
    , newPassword = ""
    , adminChecked = False
    , id = flags.id
    , accessToken = flags.accessToken
    , name = flags.name
    , apps = flags.apps
    , admin = flags.admin
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
        , ("id", Encode.string (String.fromInt model.id))
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
        [ ("new_password", Encode.string model.changePasswordText)
        , ("old_password", Encode.string model.oldPasswordText)
        , ("id", Encode.string (String.fromInt model.id))
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
            [ ("username", Encode.string model.newUserUsernameText)
            , ("password", Encode.string model.newUserPasswordText)
            , ("id", Encode.string (String.fromInt model.id))
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
        , ("id", Encode.string (String.fromInt model.id))
        ]
        
        
-- UPDATE


type Msg
  = LinkClicked Browser.UrlRequest
  | UrlChanged Url.Url
  | OtherUsernameInput String
  | NewUserUsernameInput String
  | NewUserPasswordInput String
  | ToggleAdmin
  | NewUser
  | AdminNewPassword
  | MakeAdmin
  | RevokeAdmin
  | DeleteUser
  | ChangeUsernameInput String
  | ChangePasswordInput String
  | OldPasswordInput String
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

    OldPasswordInput password ->
        ( { model | oldPasswordText = password }, Cmd.none )

    ChangePasswordInput password ->
        ( { model | changePasswordText = password }, Cmd.none )

    SubmitChangeUsername ->
        ( model, postChangeUsername model )

    SubmitChangePassword ->
        ( model, postChangePassword model )

    OtherUsernameInput username ->
        ( { model | otherUsernameText = username }, Cmd.none )

    NewUserUsernameInput username ->
        ( { model | newUserUsernameText = username }, Cmd.none )            

    NewUserPasswordInput password ->
        ( { model | newUserPasswordText = password }, Cmd.none )

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
  let
      route = parseRoute model.url
  in
      { title = "Portal"
      , body = [ text route
               , viewRouter model route                    
               ]
      }


parseRoute : Url.Url -> String
parseRoute url =
    url
    |> Url.toString
    |> String.split "/"
    |> List.drop 3
    |> List.head
    |> Maybe.withDefault "error"

               
viewRouter : Model -> String -> Html Msg
viewRouter model route =
    case route of
        "settings" ->
            settingsView model

        _ ->
            welcomeView model


viewLink : String -> Html Msg
viewLink path =
  a [ href path ] [ text path ]


welcomeView : Model -> Html Msg
welcomeView model =
    table []
        [ tr [] ((td [] [ viewLink "/settings"] ) :: List.map appView model.apps) ]

            
appView : String -> Html Msg
appView name =
    td [] [ a [ href ("/" ++ name) ] [ text name ] ]

        
settingsView : Model -> Html Msg
settingsView model =
    if model.admin then
        div [ class "setting" ]
            [ changeUsernameView model
            , changePasswordView model
            , adminSettingsView model
            , newUserView model                
            ]
    else
        div [ class "setting" ]
            [ changeUsernameView model
            , changePasswordView model
            ]

        
changeUsernameView : Model -> Html Msg
changeUsernameView model =
    div [ class "setting" ]
        [ div [ class "setting" ]
              [ text "Change Username: "
              , input [ onInput ChangeUsernameInput, placeholder "My New Username", value model.changeUsernameText ] []
              ]
        , button [ onClick SubmitChangeUsername ] [ text "Submit" ]
        ]

        
changePasswordView : Model -> Html Msg
changePasswordView model =
    div [ class "setting" ]
        [ div [ class "setting" ] [ text "Change Your password:" ]
        , div [ class "setting" ]
            [ div [ class "setting" ] [ input [ onInput OldPasswordInput, placeholder "Old Password", value model.oldPasswordText ] [] ]
            , div [ class "setting" ] [ input [ onInput ChangePasswordInput, placeholder "New Password", value model.changePasswordText ] [] ]
            , button [ onClick SubmitChangePassword ] [ text "Submit" ]
            ]
        ]


adminSettingsView : Model -> Html Msg
adminSettingsView model =
    div [ class "setting" ]
        [ div [] [ text "Update other user's settings or status:" ]
        , div []
            [ div [ class "setting" ]
                  [ text "Username: "
                  , input [ onInput OtherUsernameInput, placeholder "Other Username", value model.otherUsernameText ] []
                  ]
            , div [ class "setting" ]
                [ text "New Password: "
                , button [ onClick AdminNewPassword ] [ text "New Password" ]
                , text model.newPassword
                ]
            , div [ class "setting" ]
                [ text "Admin: "
                , button [ onClick MakeAdmin ] [ text "Yes" ]
                , button [ onClick RevokeAdmin ] [ text "No" ]
                ]
            , div [ class "setting" ]
                [ text "Delete: "
                , button [ onClick DeleteUser ] [ text "Delete User" ]
                , text "WARNING: This will delete users access to Portal."
                ]
            ]
        ]


newUserView : Model -> Html Msg
newUserView model =
    div [ class "setting" ]
        [ text "Register New User: "
        , input [ onInput NewUserUsernameInput, placeholder "Other Username", value model.newUserUsernameText ] []
        , input [ onInput NewUserPasswordInput, placeholder "New User's Password", value model.newUserPasswordText ] []
        , div [ class "setting" ]
            [ text "Admin Yes / No: "
            , input [ type_ "checkbox", checked model.adminChecked, onClick ToggleAdmin ] []
            ]
        , div [ class "setting" ] [ button [ onClick NewUser ] [ text "Register" ] ]
        ]
            
